/**
 * Kitchen Service E2E Test
 *
 * Tests the critical kitchen inventory management flow:
 * 1. Set inventory stock levels
 * 2. Query inventory
 * 3. Update inventory
 * 4. Add stock to existing items
 *
 * Note: Orders flow through Kafka, so we only test the REST inventory APIs.
 */

import { expect, test, describe, beforeAll, afterAll } from "bun:test";
import { createServiceClients, type ServiceClients } from "../utils/service-clients";
import { generateTestRunId } from "../utils/utils";
import { config } from "@config/test-config";

describe("Kitchen Service Flow", () => {
  let clients: ServiceClients;
  let testRunId: string;

  beforeAll(() => {
    testRunId = generateTestRunId();
    console.log(`Starting Kitchen Flow tests - Run ID: ${testRunId}`);
    clients = createServiceClients();
  });

  afterAll(() => {
    console.log(`Kitchen Flow tests completed - Run ID: ${testRunId}`);
  });

  test(
    "should set and retrieve inventory stock",
    async () => {
      const menuItemId = `e2e-burger-${testRunId.slice(0, 8)}`;

      // Set stock level
      const result = await clients.kitchen.updateStock({
        menuItemId,
        quantity: 50,
      });

      expect(result.menuItemId).toBe(menuItemId);
      expect(result.currentStock).toBe(50);

      // Retrieve stock level
      const stock = await clients.kitchen.getStock(menuItemId);
      expect(stock.menuItemId).toBe(menuItemId);
      expect(stock.currentStock).toBe(50);
    },
    config.timeouts.default
  );

  test(
    "should list all inventory items",
    async () => {
      // First add an item to ensure there's at least one
      const menuItemId = `e2e-fries-${testRunId.slice(0, 8)}`;
      await clients.kitchen.updateStock({
        menuItemId,
        quantity: 100,
      });

      // List all inventory
      const inventory = await clients.kitchen.listInventory();

      expect(Array.isArray(inventory)).toBe(true);
      expect(inventory.length).toBeGreaterThan(0);

      // Find our item
      const ourItem = inventory.find((item) => item.menuItemId === menuItemId);
      expect(ourItem).toBeDefined();
      expect(ourItem!.currentStock).toBe(100);
    },
    config.timeouts.default
  );

  test(
    "should add stock to existing item",
    async () => {
      const menuItemId = `e2e-pizza-${testRunId.slice(0, 8)}`;

      // Set initial stock
      await clients.kitchen.updateStock({
        menuItemId,
        quantity: 20,
      });

      // Add more stock
      const result = await clients.kitchen.addStock(menuItemId, {
        menuItemId,
        quantity: 15,
      });

      expect(result.menuItemId).toBe(menuItemId);
      expect(result.currentStock).toBe(35); // 20 + 15

      // Verify final stock level
      const stock = await clients.kitchen.getStock(menuItemId);
      expect(stock.currentStock).toBe(35);
    },
    config.timeouts.default
  );

  test(
    "should list orders (may be empty if no Kafka orders)",
    async () => {
      // List orders - this tests the endpoint is accessible
      // Orders come via Kafka, so this may be empty
      const orders = await clients.kitchen.getOrders();

      expect(Array.isArray(orders)).toBe(true);
    },
    config.timeouts.default
  );
});
