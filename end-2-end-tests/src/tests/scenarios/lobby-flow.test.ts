/**
 * Lobby Service E2E Test
 *
 * Tests the critical lobby service flows:
 * - Check-in with payment validation
 * - Check-out with payment validation
 * - Integration with Reservations, Billing, and Cleaning services
 */

import { expect, test, describe, beforeAll, afterAll } from "bun:test";
import {
  createServiceClients,
  type ServiceClients,
} from "../utils/service-clients";
import {
  generateTestRunId,
  generateGuestId,
  createTestReservation,
} from "../utils/utils";
import { config } from "@config/test-config";

describe("Lobby Service Flow", () => {
  let clients: ServiceClients;
  let testRunId: string;

  beforeAll(() => {
    testRunId = generateTestRunId();
    console.log(`Starting Lobby Flow tests - Run ID: ${testRunId}`);
    clients = createServiceClients();
  });

  afterAll(() => {
    console.log(`Lobby Flow tests completed - Run ID: ${testRunId}`);
  });

  test(
    "Test 1.1: should complete successful check-in flow with payment verification",
    async () => {
      // Setup: Create room and reservation
      const { room, folder, reservation } = await createTestReservation(
        clients,
        testRunId
      );

      console.log(
        `[Test 1.1] Created test setup - Room: ${room.number}, Guest: ${reservation.guestId}, Folder: ${folder.id}`
      );

      // Pay for the reservation
      const ticket = reservation.reservationBillingTicket;
      await clients.billing.makePayment({
        payment_method: "Cash",
        tickets: [ticket],
      });

      // Execute: Check-in
      const checkInResponse = await clients.lobby.checkIn(room.number, {
        guest_id: reservation.guestId,
        ignore_unpaid_tickets: false,
      });

      // Verify response
      expect(checkInResponse.status).toBe("success");
      expect(checkInResponse.checkin_id).toBeDefined();
      expect(checkInResponse.room_number).toBe(room.number);
      expect(checkInResponse.guest_id).toBe(reservation.guestId);
      expect(checkInResponse.timestamp).toBeDefined();

      // Verify payment status was checked
      expect(checkInResponse.payment_status).toBeDefined();
      expect(checkInResponse.payment_status.has_unpaid_tickets).toBe(false);
      expect(checkInResponse.payment_status.unpaid_tickets).toEqual([]);

      console.log(
        `[Test 1.1] Check-in successful - ID: ${checkInResponse.checkin_id}`
      );
    },
    config.timeouts.default
  );

  test(
    "Test 1.2: should block check-in when guest has unpaid tickets",
    async () => {
      // Setup: Create reservation with unpaid ticket
      const { room, folder, reservation } = await createTestReservation(
        clients,
        testRunId
      );

      console.log(
        `[Test 1.2] Created test setup - Room: ${room.number}, Guest: ${reservation.guestId}, Folder: ${folder.id}`
      );

      // Execute & Verify: Attempt check-in should fail
      try {
        await clients.lobby.checkIn(room.number, {
          guest_id: reservation.guestId,
          ignore_unpaid_tickets: false,
        });

        // Should not reach here
        expect(true).toBe(false);
      } catch (error: any) {
        // Verify error response
        expect(error.status).toBe(402); // Payment Required

        // Parse error details
        const errorDetail = error.details ? JSON.parse(error.details) : {};

        // Verify unpaid ticket information is in response
        if (typeof errorDetail.detail === "object") {
          expect(errorDetail.detail.message).toContain("unpaid");
          expect(errorDetail.detail.unpaid_tickets).toBeDefined();
          expect(errorDetail.detail.total_unpaid_amount).toBeGreaterThan(0);
        }
      }

      console.log(
        `[Test 1.2] Check-in correctly blocked due to unpaid tickets`
      );
    },
    config.timeouts.default
  );

  test(
    "Test 1.3: should complete successful check-out flow",
    async () => {
      // Setup: Create reservation and check-in first
      const { room, folder, reservation } = await createTestReservation(
        clients,
        testRunId
      );

      // Pay for the reservation
      const ticket = reservation.reservationBillingTicket;
      await clients.billing.makePayment({
        payment_method: "Cash",
        tickets: [ticket],
      });

      const checkInResponse = await clients.lobby.checkIn(room.number, {
        guest_id: reservation.guestId,
      });

      console.log(
        `[Test 1.3] Checked in - ID: ${checkInResponse.checkin_id}, now attempting check-out`
      );

      // Wait a moment to ensure check-in is processed
      await new Promise((resolve) => setTimeout(resolve, 1000));

      // Execute: Check-out
      const checkOutResponse = await clients.lobby.checkOut(
        room.number,
        {
          ignore_unpaid_tickets: false,
        }
      );

      // Verify response
      expect(checkOutResponse.status).toBe("success");
      expect(checkOutResponse.checkout_id).toBeDefined();
      expect(checkOutResponse.room_number).toBe(room.number);
      expect(checkOutResponse.guest_id).toBe(reservation.guestId);
      expect(checkOutResponse.timestamp).toBeDefined();

      // Verify payment status was checked
      expect(checkOutResponse.payment_status).toBeDefined();
      expect(checkOutResponse.payment_status.has_unpaid_tickets).toBe(false);

      console.log(
        `[Test 1.3] Check-out successful - ID: ${checkOutResponse.checkout_id}`
      );
    },
    config.timeouts.default * 2
  );

  test(
    "Test 1.4: should block check-out when guest has unpaid tickets",
    async () => {
      // Setup: Check-in guest, then create unpaid ticket during stay
      const { room, folder, reservation } = await createTestReservation(
        clients,
        testRunId
      );

      // Pay for the reservation
      const ticket = reservation.reservationBillingTicket;
      await clients.billing.makePayment({
        payment_method: "Cash",
        tickets: [ticket],
      });

      await clients.lobby.checkIn(room.number, {
        guest_id: reservation.guestId,
      });

      // Create unpaid ticket during the stay (e.g., room service)
      const unpaidTicket = await clients.billing.createTicket({
        folder_id: folder.id,
        total: "75.50",
        item_id: `room-service-${testRunId}`,
        description: "Room service order - unpaid",
      });

      await new Promise((resolve) => setTimeout(resolve, 1000));

      // Execute & Verify: Attempt check-out should fail
      try {
        await clients.lobby.checkOut(room.number, {
          ignore_unpaid_tickets: false,
        });

        // Should not reach here
        expect(true).toBe(false);
      } catch (error: any) {
        // Verify error response
        expect(error.status).toBe(402); // Payment Required

        const errorDetail = error.details ? JSON.parse(error.details) : {};

        if (typeof errorDetail.detail === "object") {
          expect(errorDetail.detail.message).toContain("unpaid");
          expect(errorDetail.detail.unpaid_tickets).toBeDefined();
        }
      }

      console.log(
        `[Test 1.4] Check-out correctly blocked due to unpaid tickets`
      );
    },
    config.timeouts.default * 2
  );

  test(
    "Test 1.5: should fail check-out for unoccupied room",
    async () => {
      // Setup: Create room but don't check in
      const { room } = await createTestReservation(clients, testRunId);

      console.log(
        `[Test 1.5] Room ${room.number} created but not checked in, attempting check-out`
      );

      // Execute & Verify: Check-out should fail
      try {
        await clients.lobby.checkOut(room.number);

        // Should not reach here
        expect(true).toBe(false);
      } catch (error: any) {
        // Verify error response
        expect(error.status).toBe(404);

        console.log(
          `[Test 1.5] Check-out correctly failed for unoccupied room`
        );
      }
    },
    config.timeouts.default
  );
});
