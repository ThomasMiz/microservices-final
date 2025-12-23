/**
 * Lobby & Cleaning Integration E2E Test
 *
 * Tests the integration between Lobby and Cleaning services:
 * - End-to-end guest journey (check-in → check-out → cleaning)
 * - Cross-service payment validation
 * - Complete workflow verification
 */

import { expect, test, describe, beforeAll, afterAll } from "bun:test";
import {
  createServiceClients,
  type ServiceClients,
} from "../utils/service-clients";
import {
  generateTestRunId,
  createTestReservation,
  waitForCleaningJob,
} from "../utils/utils";
import { config } from "@config/test-config";

describe("Lobby & Cleaning Integration", () => {
  let clients: ServiceClients;
  let testRunId: string;

  beforeAll(() => {
    testRunId = generateTestRunId();
    console.log(
      `Starting Lobby & Cleaning Integration tests - Run ID: ${testRunId}`
    );
    clients = createServiceClients();
  });

  afterAll(() => {
    console.log(
      `Lobby & Cleaning Integration tests completed - Run ID: ${testRunId}`
    );
  });

  test(
    "Test 3.1: should complete end-to-end guest journey with automatic cleaning job creation",
    async () => {
      console.log(`[Test 3.1] Starting end-to-end guest journey test`);

      // STEP 1: Setup - Create room, billing folder, and reservation
      const { room, folder, reservation } = await createTestReservation(
        clients,
        testRunId
      );

      console.log(
        `[Test 3.1] Setup complete - Room: ${room.number}, Guest: ${reservation.guestId}, Folder: ${folder.id}`
      );

      // STEP 2: Guest pays and checks in
      
      const paymentResponse = await clients.billing.makePayment({
        payment_method: "Cash",
        description: `E2E test - Payment for check-in ${testRunId}`,
        tickets: [reservation.reservationBillingTicket]
      });

      const checkInResponse = await clients.lobby.checkIn(room.number, {
        guest_id: reservation.guestId,
      });

      expect(checkInResponse.status).toBe("success");
      expect(checkInResponse.checkin_id).toBeDefined();
      console.log(
        `[Test 3.1] ✓ Check-in successful - ID: ${checkInResponse.checkin_id}`
      );

      // Wait for check-in to be processed
      await new Promise((resolve) => setTimeout(resolve, 2000));

      // STEP 3: Verify job creation
      const cleaningJob = await waitForCleaningJob(
        clients,
        room.number,
        20,
        1500
      );

      // STEP 4: Guest checks out
      const checkOutResponse = await clients.lobby.checkOut(
        room.number
      );

      expect(checkOutResponse.status).toBe("success");
      expect(checkOutResponse.checkout_id).toBeDefined();
      console.log(
        `[Test 3.1] ✓ Check-out successful - ID: ${checkOutResponse.checkout_id}`
      );

      // Verify payment status at check-out
      expect(checkOutResponse.payment_status.has_unpaid_tickets).toBe(false);
      console.log(
        `[Test 3.1] ✓ Payment status verified at check-out - No unpaid tickets`
      );

      // STEP 5: Wait for automatic cleaning job creation
      console.log(
        `[Test 3.1] Waiting for automatic cleaning job creation via Kafka...`
      );

      const cleaningJob2 = await waitForCleaningJob(
        clients,
        room.number,
        20,
        1500
      );

      // Verify cleaning job was automatically created
      expect(cleaningJob2).not.toBeNil();
      expect(cleaningJob2.roomNumber).toBe(room.number);
      expect(cleaningJob2.assignedTo).toBeNil(); // Should be unassigned
      expect(cleaningJob2.startedAt).toBeNil();
      expect(cleaningJob2.finishedAt).toBeNil();
      console.log(
        `[Test 3.1] ✓ Cleaning job automatically created - Job ID: ${cleaningJob2.id}`
      );
    },
    config.timeouts.default * 4
  );

});
