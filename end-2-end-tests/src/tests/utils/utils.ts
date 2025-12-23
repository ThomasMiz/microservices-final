/**
 * Kubernetes Test Utilities
 * 
 * Utilities for Kubernetes operations during E2E tests.
 * Includes health checks and environment management.
 */

import { config } from '@config/test-config';
import { ServiceClients } from './service-clients';

// ============================================================================
// Environment Information
// ============================================================================

export interface EnvironmentInfo {
  baseUrl: string;
  timestamp: string;
}

export function getEnvironmentInfo(): EnvironmentInfo {
  return {
    baseUrl: config.baseUrl,
    timestamp: new Date().toISOString(),
  };
}

// ============================================================================
// Test Run Utilities
// ============================================================================

/**
 * Generate a test run ID with pipeline and multirepo context
 */
export function generateTestRunId(): string {
  const timestamp = Date.now().toString(36);
  const random = Math.random().toString(36).substring(2, 8);
  const pipelineId = process.env.CI_PIPELINE_ID || 'local';
  const commitSha = process.env.CI_COMMIT_SHORT_SHA || 'dev';
  const serviceChanged = process.env.SERVICE_CHANGED || 'unknown';
  const serviceVersion = process.env.SERVICE_VERSION || 'latest';
  return `e2e-${timestamp}-${random}-pipeline-${pipelineId}-service-${serviceChanged}-v${serviceVersion}-commit-${commitSha}`;
}

/**
 * Generate a test guest name with unique identifier
 * Includes pipeline and multirepo service information for easy identification
 */
export function generateTestGuestName(testRunId: string): string {
  const pipelineId = process.env.CI_PIPELINE_ID || 'local';
  const commitSha = process.env.CI_COMMIT_SHORT_SHA || 'dev';
  const serviceChanged = process.env.SERVICE_CHANGED || 'unknown';
  const serviceVersion = process.env.SERVICE_VERSION || 'latest';
  return `e2e-test-guest-${testRunId}-pipeline-${pipelineId}-service-${serviceChanged}-v${serviceVersion}-commit-${commitSha}`;
}

/**
 * Generate a test room number (uses high numbers to avoid conflicts)
 * Includes pipeline and multirepo service context for traceability
 */
export function generateTestRoomNumber(): string {
  const floor = 9; // Use floor 9 for test rooms
  const room = Math.floor(Math.random() * 99) + 1;
  const timestamp = Date.now().toString(36).slice(-4); // Add short timestamp for uniqueness
  const pipelineId = process.env.CI_PIPELINE_ID || 'local';
  const serviceChanged = process.env.SERVICE_CHANGED || 'unknown';
  const serviceVersion = process.env.SERVICE_VERSION || 'latest';
  return `${floor}${room.toString().padStart(2, '0')}-${timestamp}-p${pipelineId}-${serviceChanged}-v${serviceVersion}`;
}

// ============================================================================
// Lobby & Cleaning Test Utilities
// ============================================================================

/**
 * Generate a unique room number for test isolation
 */
export function generateRoomNumber(): string {
  return `R${Date.now().toString().slice(-6)}`;
}

/**
 * Generate a unique guest ID for test isolation
 */
export function generateGuestId(): string {
  return `guest-${Date.now()}-${Math.random().toString(36).substring(7)}`;
}

/**
 * Wait for a cleaning job to be created for a specific room
 * Uses polling with exponential backoff
 * 
 * @param clients Service clients
 * @param roomNumber Room number to check
 * @param maxAttempts Maximum number of polling attempts
 * @param initialDelay Initial delay in ms
 * @returns The created cleaning job or null if not found
 */
export async function waitForCleaningJob(
  clients: any,
  roomNumber: string,
  maxAttempts: number = 10,
  initialDelay: number = 500
): Promise<any | null> {
  for (let attempt = 0; attempt < maxAttempts; attempt++) {
    try {
      const jobs = await clients.cleaning.searchJobs({ roomNumber });
      if (jobs.length > 0) {
        return jobs[0]; // Return the first/latest job
      }
    } catch (error) {
      // Ignore errors and continue polling
    }
    
    // Exponential backoff
    const delay = initialDelay * Math.pow(1.5, attempt);
    await new Promise(resolve => setTimeout(resolve, delay));
  }
  
  return null;
}

/**
 * Create a complete test reservation setup
 * Creates room, billing folder, and reservation
 * 
 * @param clients Service clients
 * @param testRunId Test run ID for unique naming
 * @returns Object containing room, folder, and reservation
 */
export async function createTestReservation(
  clients: ServiceClients,
  testRunId: string
) {
  const roomNumber = generateRoomNumber();
  const guestId = generateGuestId();

  // Create room
  const room = await clients.reservations.createRoom({
    number: roomNumber,
    active: true,
    name: `Test Room ${roomNumber}`,
    description: `E2E test room for run ${testRunId}`,
    maxCapacity: 2,
    category: "STANDARD",
    hourlyPrice: "50.00",
  });

  // Create reservation
  const now = new Date();
  const startDate = new Date(now.getTime() - 1000 * 60 * 60); // 1 hour ago
  const endDate = new Date(now.getTime() + 1000 * 60 * 60 * 24); // 24 hours from now

  const reservation = await clients.reservations.createReservation({
    roomId: room.id,
    guestId,
    startDate: startDate.toISOString(),
    endDate: endDate.toISOString(),
  });

  // Get folder
  const folder = await clients.billing.getFolder(reservation.billingFolderId);

  return { room, reservation, folder };
}

export default {
  getEnvironmentInfo,
  generateTestRunId,
  generateTestGuestName,
  generateTestRoomNumber,
  generateRoomNumber,
  generateGuestId,
  waitForCleaningJob,
  createTestReservation,
};
