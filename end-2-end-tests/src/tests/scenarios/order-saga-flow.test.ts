/**
 * Order Saga Flow E2E Tests
 *
 * Tests the complete order saga flow involving:
 * 1. Room Service (HTTP API - creates order, billing ticket)
 * 2. Billing Service (HTTP API - manages billing)
 * 3. Kitchen Service (Kafka consumer - processes orders)
 * 4. Reservations Service (provides room/reservation data)
 *
 * Saga Flow:
 * - Room Service receives HTTP order request
 * - Validates room has active reservation with billing folder
 * - Creates billing ticket via Billing Service API
 * - Saves order to PostgreSQL (with rollback if fails)
 * - Publishes OrderRequested event to Kafka
 * - Kitchen Service validates stock and responds (OrderAccepted/OrderRejected)
 * - If accepted, Room Service publishes OrderConfirmed after billing
 * - Kitchen transitions order through: preparing -> completed
 */

import { describe, it, expect, beforeAll } from "bun:test";
import { createServiceClients } from "../utils/service-clients";
import { generateTestRoomNumber, generateTestGuestName } from "../utils/utils";
import type {
  Room,
  Reservation,
  RoomServiceOrder,
  MenuItem,
  BillingFolder,
} from "@/types/test-types";

const clients = createServiceClients();

// Helper function to poll for order status changes
async function waitForOrderStatus(
  orderId: string,
  expectedStatus: string,
  maxAttempts = 20,
  delayMs = 1000
): Promise<RoomServiceOrder> {
  for (let i = 0; i < maxAttempts; i++) {
    const order = await clients.roomservice.getOrder(orderId);
    if (order.status === expectedStatus) {
      return order;
    }
    await new Promise((resolve) => setTimeout(resolve, delayMs));
  }
  throw new Error(
    `Order ${orderId} did not reach status ${expectedStatus} within ${maxAttempts * delayMs}ms`
  );
}

describe("Order Saga Flow E2E Tests", () => {
  let testRoom: Room;
  let testReservation: Reservation;
  let testMenuItem1: MenuItem;
  let testMenuItem2: MenuItem;
  let billingFolder: BillingFolder;

  beforeAll(async () => {
    // Setup: Create room
    testRoom = await clients.reservations.createRoom({
      number: generateTestRoomNumber(),
      name: "Order Saga Test Room",
      description: "Room for testing order saga",
      maxCapacity: 2,
      category: "STANDARD",
      hourlyPrice: "50.00",
      active: true,
    });

    // Setup: Create reservation with billing folder
    // Start reservation 5 seconds in the past to ensure it's active "now"
    const startDate = new Date(Date.now() - 5000); // 5 seconds ago
    const endDate = new Date(startDate.getTime() + 24 * 60 * 60 * 1000); // +24 hours

    testReservation = await clients.reservations.createReservation({
      roomId: testRoom.id,
      guestId: generateTestGuestName(`order-saga-${Date.now()}`),
      startDate: startDate.toISOString(),
      endDate: endDate.toISOString(),
    });

    // Get the billing folder created by reservation
    billingFolder = await clients.billing.getFolder(
      testReservation.billingFolderId
    );

    // Setup: Create menu items
    testMenuItem1 = await clients.roomservice.createMenuItem({
      name: `Saga Burger ${Date.now()}`,
      price: "15.99",
      category: "Main Course",
      available: true,
      description: "Test burger for saga",
    });

    testMenuItem2 = await clients.roomservice.createMenuItem({
      name: `Saga Fries ${Date.now()}`,
      price: "5.99",
      category: "Sides",
      available: true,
      description: "Test fries for saga",
    });

    // Setup: Set kitchen inventory
    await clients.kitchen.updateStock({
      menuItemId: testMenuItem1.ID,
      quantity: 100,
    });
    await clients.kitchen.updateStock({
      menuItemId: testMenuItem2.ID,
      quantity: 100,
    });
  });

  it("should complete order saga successfully with sufficient stock", async () => {
    // Create order via Room Service
    const orderRequest = {
      room_id: testRoom.id,
      items: [
        {
          menu_item_id: testMenuItem1.ID,
          quantity: 2,
          unit_price: testMenuItem1.Price,
        },
        {
          menu_item_id: testMenuItem2.ID,
          quantity: 3,
          unit_price: testMenuItem2.Price,
        },
      ],
    };

    const createdOrder = await clients.roomservice.createOrder(orderRequest);

    // Assert: Order created with initial status
    expect(createdOrder).toBeDefined();
    expect(createdOrder.id).toBeDefined();
    expect(createdOrder.room_id).toBe(testRoom.id);
    expect(createdOrder.reservation_id).toBe(
      testReservation.id
    );
    expect(createdOrder.billing_folder_id).toBe(
      testReservation.billingFolderId
    );
    expect(createdOrder.status).toBe("pending_kitchen");
    expect(createdOrder.items).toHaveLength(2);

    // Assert: Billing ticket will be created asynchronously
    // Initially billing_ticket_id will be empty, will be filled when billing completes
    expect(createdOrder.billing_ticket_id).toBe(""); // Initially empty
    expect(createdOrder.billing_folder_id).toBe(
      testReservation.billingFolderId
    );

    // Poll: Wait for Kitchen to accept and billing to complete (status: preparing)
    // The order goes through: pending_kitchen -> pending_billing -> preparing
    // pending_billing is transient, so we wait for preparing which indicates billing succeeded
    const preparingOrder = await waitForOrderStatus(
      createdOrder.id,
      "preparing",
      20,
      1500
    );
    expect(preparingOrder.status).toBe("preparing");
    
    // Now verify billing ticket was created
    expect(preparingOrder.billing_ticket_id).toBeDefined();
    expect(preparingOrder.billing_ticket_id).not.toBe("");
    const billingTicket = await clients.billing.getTicket(
      preparingOrder.billing_ticket_id
    );
    expect(billingTicket.state).toBe("Pending");
    expect(billingTicket.folder.id).toBe(
      testReservation.billingFolderId
    );
    
    // Verify: Kitchen order is now preparing
    const kitchenPreparingOrder = await clients.kitchen.getOrder(
      createdOrder.id
    );
    expect(kitchenPreparingOrder.status).toBe("IN_PREPARATION");

    // Final verification: Check order can be retrieved by room
    const ordersByRoom = await clients.roomservice.getOrdersByRoom(
      testRoom.id.toString()
    );
    expect(ordersByRoom).toBeDefined();
    expect(ordersByRoom.length).toBeGreaterThan(0);
    const foundOrder = ordersByRoom.find((o) => o.id === createdOrder.id);
    expect(foundOrder).toBeDefined();
  });

  it("should reject order with insufficient stock and rollback", async () => {
    // Setup: Create new menu item with low stock
    const lowStockItem = await clients.roomservice.createMenuItem({
      name: `Rare Dish ${Date.now()}`,
      price: "99.99",
      category: "Special",
      available: true,
      description: "Item with insufficient stock",
    });

    // Set very low stock
    await clients.kitchen.updateStock({
      menuItemId: lowStockItem.ID,
      quantity: 2,
    });

    // Create order that exceeds available stock
    const orderRequest = {
      room_id: testRoom.id,
      items: [
        {
          menu_item_id: lowStockItem.ID,
          quantity: 10, // More than available (2)
          unit_price: lowStockItem.Price,
        },
      ],
    };

    const createdOrder = await clients.roomservice.createOrder(orderRequest);

    // Assert: Order created initially
    expect(createdOrder).toBeDefined();
    expect(createdOrder.status).toBe("pending_kitchen");
    // Billing ticket will be created asynchronously if kitchen accepts
    expect(createdOrder.billing_ticket_id).toBe(""); // Initially empty

    // Poll: Wait for Kitchen to reject the order (should happen due to insufficient stock)
    // The order should go directly from pending_kitchen -> cancelled (no billing created)
    const rejectedOrder = await waitForOrderStatus(
      createdOrder.id,
      "cancelled",
      20,
      1500
    );
    expect(rejectedOrder.status).toBe("cancelled");

    // Verify: Kitchen order shows rejection (if it exists)
    // Note: Kitchen might not have persisted the order if it was rejected during validation
    try {
      const kitchenOrder = await clients.kitchen.getOrder(createdOrder.id);
      expect(kitchenOrder.status).toBe("REJECTED");
    } catch (error: any) {
      // If 404, it means kitchen never persisted the order (rejected during validation)
      if (error.status === 404) {
        console.log("Kitchen order not found (rejected during validation) - this is expected");
      } else {
        throw error;
      }
    }

    // Verify: Inventory not deducted (should still be 2)
    const inventory = await clients.kitchen.getStock(lowStockItem.ID);
    expect(inventory.currentStock).toBe(2);
  });
});
