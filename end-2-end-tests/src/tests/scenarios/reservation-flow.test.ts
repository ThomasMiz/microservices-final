/**
 * Reservation Service E2E Test
 *
 * Tests the critical reservation flow:
 * 1. Create a room
 * 2. Search for available rooms
 * 3. Create a reservation (which creates billing folder + ticket)
 * 4. Retrieve reservation details
 */

import { expect, test, describe, beforeAll, afterAll } from "bun:test";
import { createServiceClients, type ServiceClients } from "../utils/service-clients";
import { generateTestRunId, generateTestRoomNumber, generateTestGuestName } from "../utils/utils";
import { config } from "@config/test-config";

describe("Reservation Service Flow", () => {
  let clients: ServiceClients;
  let testRunId: string;

  beforeAll(() => {
    testRunId = generateTestRunId();
    console.log(`Starting Reservation Flow tests - Run ID: ${testRunId}`);
    clients = createServiceClients();
  });

  afterAll(() => {
    console.log(`Reservation Flow tests completed - Run ID: ${testRunId}`);
  });

  test(
    "should create a room and verify it can be retrieved",
    async () => {
      // Create a new room with unique room number
      const room = await clients.reservations.createRoom({
        number: generateTestRoomNumber(),
        active: true,
        name: `E2E Test Room ${testRunId}`,
        description: "Room created for E2E testing",
        maxCapacity: 2,
        category: "STANDARD",
        hourlyPrice: "25.00",
      });

      expect(room.id).toBeDefined();
      expect(room.number).toMatch(/^9\d{2}-/); // Should start with 9 and be 4 digits
      expect(room.active).toBe(true);
      expect(room.category).toBe("STANDARD");

      // Verify room can be retrieved
      const retrievedRoom = await clients.reservations.getRoom(room.id);
      expect(retrievedRoom.id).toBe(room.id);
      expect(retrievedRoom.name).toBe(`E2E Test Room ${testRunId}`);
    },
    config.timeouts.default
  );

  test(
    "should search for available rooms",
    async () => {
      // Search for rooms
      const rooms = await clients.reservations.searchRooms();

      expect(Array.isArray(rooms)).toBe(true);
      // Should have at least one room (created in previous test or pre-existing)
    },
    config.timeouts.default
  );

  test(
    "should complete full reservation flow with billing integration",
    async () => {
      // Step 1: Create a room for this reservation
      const room = await clients.reservations.createRoom({
        number: generateTestRoomNumber(),
        active: true,
        name: `Reservation Test Room ${testRunId}`,
        description: "Room for reservation E2E test",
        maxCapacity: 4,
        category: "DELUXE",
        hourlyPrice: "50.00",
      });

      expect(room.id).toBeDefined();

      // Step 2: Create a reservation (24 hours from now, for 2 days)
      const startDate = new Date();
      startDate.setDate(startDate.getDate() + 1);
      startDate.setHours(14, 0, 0, 0);

      const endDate = new Date(startDate);
      endDate.setDate(endDate.getDate() + 2);
      endDate.setHours(11, 0, 0, 0);

      const reservation = await clients.reservations.createReservation({
        roomId: room.id,
        guestId: generateTestGuestName(testRunId),
        startDate: startDate.toISOString(),
        endDate: endDate.toISOString(),
      });

      expect(reservation.id).toBeDefined();
      expect(reservation.room.id).toBe(room.id);
      expect(reservation.guestId).toContain(`guest-${testRunId}`);
      expect(reservation.billingFolderId).toBeDefined();
      expect(reservation.reservationBillingTicket).toBeDefined();
      expect(parseFloat(reservation.totalPrice)).toBeGreaterThan(0);

      // Step 3: Verify reservation can be retrieved
      const retrievedReservation = await clients.reservations.getReservation(
        reservation.id
      );
      expect(retrievedReservation.id).toBe(reservation.id);
      expect(retrievedReservation.guestId).toContain(`guest-${testRunId}`);

      // Step 4: Verify billing folder was created
      const folder = await clients.billing.getFolder(reservation.billingFolderId);
      expect(folder.id).toBe(reservation.billingFolderId);
      expect(folder.closed_at).toBeNull();

      // Step 5: Verify billing ticket was created
      const ticket = await clients.billing.getTicket(
        reservation.reservationBillingTicket
      );
      expect(ticket.id).toBe(reservation.reservationBillingTicket);
      expect(ticket.folder.id).toBe(reservation.billingFolderId);
      expect(ticket.state).toBe("Pending");
    },
    config.timeouts.default
  );

  test(
    "should list reservations",
    async () => {
      // Search for reservations
      const reservations = await clients.reservations.searchReservations();

      expect(Array.isArray(reservations)).toBe(true);
      // Should have at least one reservation from previous test
      expect(reservations.length).toBeGreaterThan(0);
    },
    config.timeouts.default
  );
});
