/**
 * E2E Test Type Definitions
 *
 * Type definitions matching the actual microservice APIs.
 */

// ============================================================================
// Billing Service Types (Rust/Axum)
// Base path: /api/billing
// ============================================================================

export interface BillingFolder {
  id: string; // UUID
  created_at: string; // ISO datetime
  closed_at: string | null;
  name: string | null;
}

export interface CreateFolderRequest {
  name?: string;
}

export interface BillingTicket {
  id: string; // UUID
  folder: BillingFolder;
  total: string; // BigDecimal as string
  item_id: string;
  description: string | null;
  created_at: string;
  closed_at: string | null;
  state: "Pending" | "Paid" | "Canceled" | "Paying";
}

export interface CreateTicketRequest {
  folder_id: string; // UUID
  total: string; // BigDecimal as string
  item_id: string;
  description?: string;
}

export type PaymentMethod = "Cash" | "Credit" | "Debit" | "Dishwashing" | "Other";

export interface MakePaymentRequest {
  payment_method: PaymentMethod;
  description?: string;
  tickets: string[]; // Array of ticket UUIDs
}

export interface MessageResponse {
  message: string;
}

// ============================================================================
// Reservations Service Types (Java/Micronaut)
// Base path: /api/reservations
// ============================================================================

export type RoomCategory = "STANDARD" | "DELUXE" | "SUITE" | "PENTHOUSE";

export interface Room {
  id: number;
  number: string;
  active: boolean;
  name: string;
  description: string;
  maxCapacity: number;
  category: RoomCategory;
  hourlyPrice: string; // BigDecimal as string
}

export interface RoomWithAvailability extends Room {
  availableRanges: DateRange[];
}

export interface DateRange {
  start: string; // ISO datetime
  end: string;
}

export interface CreateRoomRequest {
  number: string;
  active?: boolean;
  name: string;
  description: string;
  maxCapacity: number;
  category: RoomCategory;
  hourlyPrice: string;
}

export interface Reservation {
  id: number;
  room: Room;
  guestId: string;
  startDate: string; // ISO datetime
  endDate: string;
  billingFolderId: string;
  reservationBillingTicket: string;
  rentedHourlyPrice: string;
  totalPrice: string;
}

export interface CreateReservationRequest {
  roomId: number;
  guestId: string;
  startDate: string; // ISO datetime
  endDate: string;
}

// ============================================================================
// Kitchen Service Types (Java/Quarkus)
// Base path: /api/kitchen
// ============================================================================

export interface InventoryItem {
  menuItemId: string;
  quantity: number;
}

export interface InventoryResponse {
  menuItemId: string;
  currentStock: number;
  lastUpdated: string;
}

export interface InventoryUpdateRequest {
  menuItemId: string;
  quantity: number;
}

export interface KitchenOrder {
  id: string;
  roomId: string;
  reservationId: string;
  items: OrderItem[];
  totalPrice: string;
  status: string;
  createdAt: string;
}

export interface OrderItem {
  menuItemId: string;
  quantity: number;
  unitPrice?: string;
}

export interface OrderStatusUpdate {
  status: "COMPLETED" | "CANCELLED";
}

// ============================================================================
// Room Service Types (Go/Gin)
// Base path: /api/roomservice
// ============================================================================

export interface MenuItem {
  ID: string;
  Description: string;
  Price: string; // decimal.Decimal as string
  CreatedAt: string;
  UpdatedAt: string;
  // Note: Name, Category, Available are not returned by the API
  // Only sent in CreateMenuItemRequest
}

export interface CreateMenuItemRequest {
  name: string;
  price: string; // decimal.Decimal as string
  category: string;
  available: boolean;
  description: string;
}

export type OrderStatus =
  | "pending_kitchen"
  | "pending_billing"
  | "preparing"
  | "completed"
  | "cancelled";

export interface RoomServiceOrder {
  id: string;
  room_id: number;
  reservation_id: number;
  billing_folder_id: string;
  billing_ticket_id: string;
  items: RoomServiceOrderItem[];
  total_price: string; // decimal.Decimal as string
  status: OrderStatus;
  created_at: string; // ISO datetime
  updated_at: string;
}

export interface RoomServiceOrderItem {
  menu_item_id: string;
  quantity: number;
  unit_price: string; // decimal.Decimal as string
}

export interface CreateOrderRequest {
  room_id: number;
  items: RoomServiceOrderItem[];
}

// ============================================================================
// Lobby Service Types (Python/FastAPI)
// Base path: /api/lobby or / (configurable via BASE_PATH)
// ============================================================================

export interface CheckInRequest {
  guest_id?: string;
  ignore_unpaid_tickets?: boolean;
}

export interface CheckInResponse {
  status: string;
  checkin_id: string;
  room_number: string;
  guest_id: string;
  timestamp: string;
  payment_status: PaymentStatus;
}

export interface CheckOutRequest {
  ignore_unpaid_tickets?: boolean;
}

export interface CheckOutResponse {
  status: string;
  message: string;
  checkout_id: string;
  room_number: string;
  guest_id: string;
  timestamp: string;
  payment_status: PaymentStatus;
}

export interface PaymentStatus {
  has_unpaid_tickets: boolean;
  unpaid_tickets: UnpaidTicket[];
  total_unpaid_amount: number;
  payment_check_skipped?: boolean;
  error?: string;
}

export interface UnpaidTicket {
  ticket_id: string;
  total: number;
  description: string;
  item_id: string;
  folder_id: string;
}

export interface LobbyErrorResponse {
  detail: {
    message: string;
    unpaid_tickets: UnpaidTicket[];
    total_unpaid_amount: number;
  } | string;
}

// ============================================================================
// Cleaning Service Types (Java/Micronaut)
// Base path: /api/cleaning
// Note: Cleaning service requires authentication
// ============================================================================

export interface CleaningJob {
  id: number;
  roomNumber: string;
  assignedTo: CleaningStaff | null;
  createdAt: string;
  startedAt: string | null;
  finishedAt: string | null;
}

export interface CleaningStaff {
  id: number;
  name: string;
  createdAt: string;
}

export interface CreateCleaningStaffRequest {
  name: string;
}

export interface AssignStaffRequest {
  staffId: number;
}

export interface DamageReport {
  id: number;
  roomNumber: string;
  reportedBy: CleaningStaff;
  item: string;
  description: string;
  guestId: string;
  fineAmount: number;
  billingFolder: string;
  billingTicket: string;
  createdAt: string;
}

export interface CreateDamageReportRequest {
  roomNumber: string;
  brokenItem: string;
  description: string;
  fineAmount: string;
}

export interface RoomOccupancyStatus {
  roomNumber: string;
  available: boolean;
  state: "OCCUPIED_DIRTY" | "OCCUPIED_CLEAN" | "EMPTY_DIRTY" | "EMPTY_CLEAN";
  currentGuestId: string | null;
  updatedAt: string;
}

export interface SearchCleaningJobsParams {
  roomNumber?: string;
  assignedTo?: number;
  page?: number;
  pageSize?: number;
}

// ============================================================================
// Test Utility Types
// ============================================================================

export interface ServiceClientOptions {
  timeout?: number;
  retries?: number;
  headers?: Record<string, string>;
}

export interface ApiResponse<T> {
  data: T;
  status: number;
  headers: Headers;
}

export interface ApiError {
  message: string;
  status: number;
  details?: unknown;
}
