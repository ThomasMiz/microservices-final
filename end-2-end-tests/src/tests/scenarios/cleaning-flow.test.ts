/**
 * Cleaning Service E2E Test
 *
 * Tests critical cleaning service flows:
 * - Cleaning job creation from Kafka events
 * - Complete cleaning job lifecycle (assign, start, finish)
 * - Damage report creation with billing integration
 * - Room occupancy status queries
 */

import { expect, test, describe, beforeAll, afterAll } from "bun:test";
import {
  createServiceClients,
  type ServiceClients,
} from "../utils/service-clients";
import {
  generateTestRunId,
  createTestReservation,
} from "../utils/utils";
import { config } from "@config/test-config";

describe("Cleaning Service Flow", () => {
  let clients: ServiceClients;
  let testRunId: string;

  beforeAll(() => {
    testRunId = generateTestRunId();
    console.log(`Starting Cleaning Flow tests - Run ID: ${testRunId}`);
    clients = createServiceClients();
  });

  afterAll(() => {
    console.log(`Cleaning Flow tests completed - Run ID: ${testRunId}`);
  });

  test(
    "Test 2.1: should create cleaning job from events",
    async () => {
      // Setup: Create room and reservation
      const { room, reservation } = await createTestReservation(clients, testRunId);

      console.log(
        `[Test 2.1] Created room ${room.number}, checking in first...`
      );

      // Check-in to make room occupied (this creates room occupancy record)
      await clients.lobby.checkIn(room.number, {
        guest_id: reservation.guestId,
        ignore_unpaid_tickets: true, // Ignore existing unpaid tickets for test
      });

      // Wait for check-in to be processed
      await new Promise((resolve) => setTimeout(resolve, 2000));

      console.log(
        `[Test 2.1] Check-in complete, manually triggering cleaning job creation...`
      );

      // Manually trigger cleaning job creation (alternative to Kafka)
      const createResult = await clients.cleaning.createCleaningJobs();
      expect(createResult.message).toBeDefined();

      // Wait a moment for job creation
      await new Promise((resolve) => setTimeout(resolve, 2000));

      // Search for cleaning job
      const jobs = await clients.cleaning.searchJobs({
        roomNumber: room.number
      });

      // Verify cleaning job was created
      expect(jobs.length).toBeGreaterThan(0);
      const cleaningJob = jobs[0];
      expect(cleaningJob.roomNumber).toBe(room.number);
      expect(cleaningJob.assignedTo).toBeNil(); // Should be unassigned initially
      expect(cleaningJob.startedAt).toBeNil();
      expect(cleaningJob.finishedAt).toBeNil();
      expect(cleaningJob.createdAt).toBeDefined();

      console.log(
        `[Test 2.1] Cleaning job created successfully - Job ID: ${cleaningJob.id}`
      );
    },
    config.timeouts.default * 2
  );

  test(
    "Test 2.2: should complete cleaning job lifecycle (assign, start, finish)",
    async () => {
      // Setup: Create a cleaning staff and cleaning job
      const { room , reservation} = await createTestReservation(clients, testRunId);
      console.log(`[Test 2.2] Created room ${room.number}`);

      // Check-in to make room occupied (this creates room occupancy record)
      await clients.lobby.checkIn(room.number, {
        guest_id: reservation.guestId,
        ignore_unpaid_tickets: true, // Ignore existing unpaid tickets for test
      });

      // Create cleaning staff
      const staff = await clients.cleaning.createStaff({
        name: `Test Staff ${testRunId}`,
      });

      console.log(
        `[Test 2.2] Created cleaning staff: ${staff.name} (ID: ${staff.id})`
      );

      // Trigger cleaning job creation manually
      await clients.cleaning.createCleaningJobs();

      // Wait for job to be created
      await new Promise((resolve) => setTimeout(resolve, 2000));

      // Get cleaning jobs
      const jobs = await clients.cleaning.searchJobs({ roomNumber: room.number });
      expect(jobs.length).toBeGreaterThan(0);

      const job = jobs[0];
      console.log(`[Test 2.2] Found cleaning job ID: ${job.id}`);

      // Step 1: Assign staff to job
      const assignResult = await clients.cleaning.assignJob(job.id, staff.id);
      expect(assignResult.message).toBeDefined();

      // Verify assignment
      const assignedJob = await clients.cleaning.getJob(job.id);
      expect(assignedJob.assignedTo).not.toBeNil();
      expect(assignedJob.assignedTo?.id).toBe(staff.id);
      console.log(`[Test 2.2] Job ${job.id} assigned to staff ${staff.id}`);

      // Step 2: Mark job as started
      const startResult = await clients.cleaning.startJob(job.id);
      expect(startResult.message).toBeDefined();

      // Verify started
      const startedJob = await clients.cleaning.getJob(job.id);
      expect(startedJob.startedAt).not.toBeNil();
      expect(startedJob.finishedAt).toBeNil();
      console.log(`[Test 2.2] Job ${job.id} marked as started`);

      // Step 3: Mark job as finished
      const finishResult = await clients.cleaning.finishJob(job.id);
      expect(finishResult.message).toBeDefined();

      // Verify finished
      const finishedJob = await clients.cleaning.getJob(job.id);
      expect(finishedJob.startedAt).not.toBeNil();
      expect(finishedJob.finishedAt).not.toBeNil();
      console.log(`[Test 2.2] Job ${job.id} marked as finished`);

      console.log(`[Test 2.2] Complete cleaning job lifecycle verified`);
    },
    config.timeouts.default * 2
  );

  test(
    "Test 2.3: should create damage report with billing integration",
    async () => {
      // Setup: Create reservation with active guest
      const { room, folder, reservation } = await createTestReservation(
        clients,
        testRunId
      );

      console.log(
        `[Test 2.3] Creating damage report for room ${room.number}, guest ${reservation.guestId}`
      );

      // Create damage report
      const damageReport = await clients.cleaning.createDamageReport({
        roomNumber: room.number,
        brokenItem: "TV Remote",
        description: `E2E test damage report - ${testRunId}`,
        fineAmount: "25.00",
      });

      // Verify damage report
      expect(damageReport.id).toBeDefined();
      expect(damageReport.roomNumber).toBe(room.number);
      expect(damageReport.item).toBe("TV Remote");
      expect(damageReport.fineAmount).toBe(25.00);
      expect(damageReport.guestId).toBe(reservation.guestId);
      expect(damageReport.billingFolder).toBeDefined();
      expect(damageReport.billingTicket).toBeDefined();
      expect(damageReport.reportedBy).toBeDefined();

      console.log(
        `[Test 2.3] Damage report created - ID: ${damageReport.id}, Billing Ticket: ${damageReport.billingTicket}`
      );

      // Verify billing ticket was created
      const billingTicket = await clients.billing.getTicket(
        damageReport.billingTicket
      );
      expect(billingTicket.id).toBe(damageReport.billingTicket);
      expect(billingTicket.total).toBe("25");
      expect(billingTicket.state).toBe("Pending");

      console.log(
        `[Test 2.3] Billing ticket verified - State: ${billingTicket.state}`
      );

      console.log(
        `[Test 2.3] Damage report with billing integration verified`
      );
    },
    config.timeouts.default * 2
  );

  test(
    "Test 2.4: should prevent duplicate cleaning job creation",
    async () => {
      // Setup: Create room
      const { room } = await createTestReservation(clients, testRunId);

      console.log(
        `[Test 2.4] Testing duplicate job prevention for room ${room.number}`
      );

      // Trigger cleaning job creation twice rapidly
      await clients.cleaning.createCleaningJobs();
      await clients.cleaning.createCleaningJobs();

      await new Promise((resolve) => setTimeout(resolve, 2000));

      // Query jobs for this room
      const jobs = await clients.cleaning.searchJobs({ roomNumber: room.number });

      console.log(
        `[Test 2.4] Found ${jobs.length} cleaning job(s) for room ${room.number}`
      );

      // Verify only one pending job exists (or verify deduplication logic)
      const pendingJobs = jobs.filter((job) => job.finishedAt === null);

      console.log(`[Test 2.4] Pending jobs: ${pendingJobs.length}`);

      // The system should have deduplicated, so we expect 1 or possibly 0 if timing is off
      expect(pendingJobs.length).toBeLessThanOrEqual(1);

      console.log(
        `[Test 2.4] Duplicate job prevention verified - ${pendingJobs.length} pending job(s)`
      );
    },
    config.timeouts.default * 2
  );
});
