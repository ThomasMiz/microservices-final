/**
 * Billing Service E2E Test
 *
 * Tests the critical billing flow:
 * 1. Create a billing folder
 * 2. Add tickets to the folder
 * 3. Pay the tickets
 * 4. Close the folder
 */

import { expect, test, describe, beforeAll, afterAll } from "bun:test";
import { createServiceClients, type ServiceClients } from "../utils/service-clients";
import { generateTestRunId } from "../utils/utils";
import { config } from "@config/test-config";

describe("Billing Service Flow", () => {
  let clients: ServiceClients;
  let testRunId: string;

  beforeAll(() => {
    testRunId = generateTestRunId();
    console.log(`Starting Billing Flow tests - Run ID: ${testRunId}`);
    clients = createServiceClients();
  });

  afterAll(() => {
    console.log(`Billing Flow tests completed - Run ID: ${testRunId}`);
  });

  test(
    "should complete full billing lifecycle: folder -> tickets -> payment -> close",
    async () => {
      // Step 1: Create a billing folder
      const folder = await clients.billing.createFolder({
        name: `e2e-test-folder-${testRunId}`,
      });

      expect(folder.id).toBeDefined();
      expect(folder.name).toBe(`e2e-test-folder-${testRunId}`);
      expect(folder.closed_at).toBeNull();

      // Step 2: Verify folder can be retrieved
      const retrievedFolder = await clients.billing.getFolder(folder.id);
      expect(retrievedFolder.id).toBe(folder.id);

      // Step 3: Create first ticket
      const ticket1 = await clients.billing.createTicket({
        folder_id: folder.id,
        total: "50.00",
        item_id: `e2e-item-1-${testRunId}`,
        description: "Room service order",
      });

      expect(ticket1.id).toBeDefined();
      expect(ticket1.folder.id).toBe(folder.id);
      expect(ticket1.total).toBe("50.00");
      expect(ticket1.state).toBe("Pending");

      // Step 4: Create second ticket
      const ticket2 = await clients.billing.createTicket({
        folder_id: folder.id,
        total: "25.50",
        item_id: `e2e-item-2-${testRunId}`,
        description: "Mini bar",
      });

      expect(ticket2.id).toBeDefined();
      expect(ticket2.state).toBe("Pending");

      // Step 5: Verify tickets can be searched by folder
      const tickets = await clients.billing.searchTickets(folder.id);
      expect(tickets.length).toBe(2);

      // Step 6: Pay all tickets
      const paymentResult = await clients.billing.makePayment({
        payment_method: "Dishwashing",
        description: `E2E test payment - ${testRunId}`,
        tickets: [ticket1.id, ticket2.id],
      });

      expect(paymentResult.message).toContain("paid");

      // Step 7: Verify tickets are now paid
      const paidTicket1 = await clients.billing.getTicket(ticket1.id);
      expect(paidTicket1.state).toBe("Paid");

      const paidTicket2 = await clients.billing.getTicket(ticket2.id);
      expect(paidTicket2.state).toBe("Paid");

      // Step 8: Close the folder
      const closeResult = await clients.billing.closeFolder(folder.id);
      expect(closeResult.message).toContain("closed");

      // Step 9: Verify folder is closed
      const closedFolder = await clients.billing.getFolder(folder.id);
      expect(closedFolder.closed_at).not.toBeNull();
    },
    config.timeouts.default
  );

  test(
    "should prevent closing folder with unpaid tickets",
    async () => {
      // Create folder with unpaid ticket
      const folder = await clients.billing.createFolder({
        name: `e2e-unpaid-test-${testRunId}`,
      });

      await clients.billing.createTicket({
        folder_id: folder.id,
        total: "100.00",
        item_id: `e2e-unpaid-item-${testRunId}`,
        description: "Unpaid item",
      });

      // Attempt to close folder with unpaid ticket - should fail
      try {
        await clients.billing.closeFolder(folder.id);
        expect(true).toBe(false); // Should not reach here
      } catch (error: unknown) {
        const apiError = error as { status: number };
        expect(apiError.status).toBe(400);
      }
    },
    config.timeouts.default
  );

  test(
    "should cancel ticket successfully",
    async () => {
      // Create folder and ticket
      const folder = await clients.billing.createFolder({
        name: `e2e-cancel-test-${testRunId}`,
      });

      const ticket = await clients.billing.createTicket({
        folder_id: folder.id,
        total: "30.00",
        item_id: `e2e-cancel-item-${testRunId}`,
        description: "Item to cancel",
      });

      expect(ticket.state).toBe("Pending");

      // Cancel the ticket
      const cancelResult = await clients.billing.deleteTicket(ticket.id);
      expect(cancelResult.message).toContain("deleted");

      // Folder can now be closed (no remaining tickets)
      const closeResult = await clients.billing.closeFolder(folder.id);
      expect(closeResult.message).toContain("closed");
    },
    config.timeouts.default
  );
});
