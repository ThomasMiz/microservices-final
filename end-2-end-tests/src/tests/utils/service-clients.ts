/**
 * Service Clients
 *
 * HTTP clients for CHotel microservices matching actual API endpoints.
 */

import { config } from "@config/test-config";
import type {
  BillingFolder,
  BillingTicket,
  CreateFolderRequest,
  CreateTicketRequest,
  MakePaymentRequest,
  MessageResponse,
  Room,
  RoomWithAvailability,
  CreateRoomRequest,
  Reservation,
  CreateReservationRequest,
  InventoryResponse,
  InventoryUpdateRequest,
  KitchenOrder,
  OrderStatusUpdate,
  MenuItem,
  CreateMenuItemRequest,
  RoomServiceOrder,
  CreateOrderRequest,
  ApiResponse,
  ApiError,
  ServiceClientOptions,
  CheckInRequest,
  CheckInResponse,
  CheckOutRequest,
  CheckOutResponse,
  CleaningJob,
  CleaningStaff,
  CreateCleaningStaffRequest,
  AssignStaffRequest,
  DamageReport,
  CreateDamageReportRequest,
  RoomOccupancyStatus,
  SearchCleaningJobsParams,
} from "@/types/test-types";

// ============================================================================
// Base HTTP Client
// ============================================================================

export class HttpClient {
  protected baseUrl: string;
  protected defaultOptions: ServiceClientOptions;

  constructor(baseUrl: string, options: ServiceClientOptions = {}) {
    this.baseUrl = baseUrl;
    this.defaultOptions = {
      timeout: options.timeout || config.timeouts.default,
      retries: options.retries || 3,
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
    };
  }

  protected async request<T>(
    method: string,
    endpoint: string,
    body?: unknown,
    options: ServiceClientOptions = {}
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseUrl}${endpoint}`;
    const mergedOptions = { ...this.defaultOptions, ...options };

    const controller = new AbortController();
    const timeoutId = setTimeout(
      () => controller.abort(),
      mergedOptions.timeout
    );

    try {
      const response = await fetch(url, {
        method,
        headers: mergedOptions.headers,
        body: body ? JSON.stringify(body) : undefined,
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        const error: ApiError = {
          message: `HTTP ${response.status}: ${response.statusText}`,
          status: response.status,
          details: await response.text().catch(() => null),
        };
        throw error;
      }

      const data = (await response.json()) as T;
      return {
        data,
        status: response.status,
        headers: response.headers,
      };
    } catch (error) {
      clearTimeout(timeoutId);

      if (error instanceof Error && error.name === "AbortError") {
        throw { message: "Request timeout", status: 408 } as ApiError;
      }
      throw error;
    }
  }

  async get<T>(
    endpoint: string,
    options?: ServiceClientOptions
  ): Promise<ApiResponse<T>> {
    return this.request<T>("GET", endpoint, undefined, options);
  }

  async post<T>(
    endpoint: string,
    body: unknown,
    options?: ServiceClientOptions
  ): Promise<ApiResponse<T>> {
    return this.request<T>("POST", endpoint, body, options);
  }

  async patch<T>(
    endpoint: string,
    body: unknown,
    options?: ServiceClientOptions
  ): Promise<ApiResponse<T>> {
    return this.request<T>("PATCH", endpoint, body, options);
  }

  async delete<T>(
    endpoint: string,
    options?: ServiceClientOptions
  ): Promise<ApiResponse<T>> {
    return this.request<T>("DELETE", endpoint, undefined, options);
  }

  async healthCheck(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/health`, {
        method: "GET",
        signal: AbortSignal.timeout(config.timeouts.default),
      });
      return response.ok;
    } catch {
      return false;
    }
  }
}

// ============================================================================
// Billing Service Client
// Endpoints: /folders, /folders/{id}, /folders/{id}/close, /tickets, /tickets/{id}, /payments
// ============================================================================

export class BillingClient extends HttpClient {
  constructor(options?: ServiceClientOptions) {
    super(config.services.billing.url, options);
  }

  // Folder operations
  async createFolder(data: CreateFolderRequest): Promise<BillingFolder> {
    const response = await this.post<BillingFolder>("/folders", data);
    return response.data;
  }

  async getFolder(id: string): Promise<BillingFolder> {
    const response = await this.get<BillingFolder>(`/folders/${id}`);
    return response.data;
  }

  async closeFolder(id: string): Promise<MessageResponse> {
    const response = await this.post<MessageResponse>(
      `/folders/${id}/close`,
      {}
    );
    return response.data;
  }

  async deleteFolder(id: string): Promise<MessageResponse> {
    const response = await this.delete<MessageResponse>(`/folders/${id}`);
    return response.data;
  }

  // Ticket operations
  async createTicket(data: CreateTicketRequest): Promise<BillingTicket> {
    const response = await this.post<BillingTicket>("/tickets", data);
    return response.data;
  }

  async getTicket(id: string): Promise<BillingTicket> {
    const response = await this.get<BillingTicket>(`/tickets/${id}`);
    return response.data;
  }

  async searchTickets(folderId?: string): Promise<BillingTicket[]> {
    const query = folderId ? `?folderId=${folderId}` : "";
    const response = await this.get<BillingTicket[]>(`/tickets${query}`);
    return response.data;
  }

  async deleteTicket(id: string): Promise<MessageResponse> {
    const response = await this.delete<MessageResponse>(`/tickets/${id}`);
    return response.data;
  }

  // Payment operations
  async makePayment(data: MakePaymentRequest): Promise<MessageResponse> {
    const response = await this.post<MessageResponse>("/payments", data);
    return response.data;
  }
}

// ============================================================================
// Reservations Service Client
// Endpoints: /rooms, /rooms/{id}, /reservations, /reservations/{id}
// ============================================================================

export class ReservationsClient extends HttpClient {
  constructor(options?: ServiceClientOptions) {
    super(config.services.reservations.url, options);
  }

  // Room operations
  async createRoom(data: CreateRoomRequest): Promise<Room> {
    const response = await this.post<Room>("/rooms", data);
    return response.data;
  }

  async getRoom(id: number): Promise<RoomWithAvailability> {
    const response = await this.get<RoomWithAvailability>(`/rooms/${id}`);
    return response.data;
  }

  async searchRooms(): Promise<RoomWithAvailability[]> {
    const response = await this.get<RoomWithAvailability[]>("/rooms");
    return response.data;
  }

  // Reservation operations
  async createReservation(data: CreateReservationRequest): Promise<Reservation> {
    const response = await this.post<Reservation>("/reservations", data);
    return response.data;
  }

  async getReservation(id: number): Promise<Reservation> {
    const response = await this.get<Reservation>(`/reservations/${id}`);
    return response.data;
  }

  async searchReservations(): Promise<Reservation[]> {
    const response = await this.get<Reservation[]>("/reservations");
    return response.data;
  }
}

// ============================================================================
// Kitchen Service Client
// Endpoints: /inventory, /inventory/{menuItemId}, /inventory/bulk, /inventory/{menuItemId}/add
//            /orders, /orders/{id}
// ============================================================================

export class KitchenClient extends HttpClient {
  constructor(options?: ServiceClientOptions) {
    super(config.services.kitchen.url, options);
  }

  // Inventory operations
  async updateStock(data: InventoryUpdateRequest): Promise<InventoryResponse> {
    const response = await this.post<InventoryResponse>("/inventory", data);
    return response.data;
  }

  async getStock(menuItemId: string): Promise<InventoryResponse> {
    const response = await this.get<InventoryResponse>(
      `/inventory/${menuItemId}`
    );
    return response.data;
  }

  async listInventory(): Promise<InventoryResponse[]> {
    const response = await this.get<InventoryResponse[]>("/inventory");
    return response.data;
  }

  async addStock(
    menuItemId: string,
    data: InventoryUpdateRequest
  ): Promise<InventoryResponse> {
    const response = await this.post<InventoryResponse>(
      `/inventory/${menuItemId}/add`,
      data
    );
    return response.data;
  }

  // Order operations (orders come via Kafka, but can be queried via REST)
  async getOrders(): Promise<KitchenOrder[]> {
    const response = await this.get<KitchenOrder[]>("/orders");
    return response.data;
  }

  async getOrder(id: string): Promise<KitchenOrder> {
    const response = await this.get<KitchenOrder>(`/orders/${id}`);
    return response.data;
  }

  async updateOrderStatus(
    id: string,
    data: OrderStatusUpdate
  ): Promise<void> {
    await this.patch<void>(`/orders/${id}`, data);
  }
}

// ============================================================================
// Room Service Client
// Endpoints: /menu, /menu/items, /orders, /orders/{id}, /orders/room/{roomId}
// ============================================================================

export class RoomServiceClient extends HttpClient {
  constructor(options?: ServiceClientOptions) {
    super(config.services.roomservice.url, options);
  }

  // Menu operations
  async createMenuItem(data: CreateMenuItemRequest): Promise<MenuItem> {
    const response = await this.post<MenuItem>("/menu/items", data);
    return response.data;
  }

  async getMenu(): Promise<MenuItem[]> {
    const response = await this.get<MenuItem[]>("/menu");
    return response.data;
  }

  // Order operations
  async createOrder(data: CreateOrderRequest): Promise<RoomServiceOrder> {
    const response = await this.post<any>("/orders", data);
    
    // Transform API response (PascalCase) to TypeScript types (snake_case)
    const apiData = response.data;
    const transformedOrder: RoomServiceOrder = {
      id: apiData.ID,
      room_id: parseInt(apiData.RoomID),
      reservation_id: parseInt(apiData.ReservationID),
      billing_folder_id: apiData.BillingFolderID,
      billing_ticket_id: apiData.BillingTicketID || "",
      items: apiData.Items.map((item: any) => ({
        menu_item_id: item.MenuItemID,
        quantity: item.Quantity,
        unit_price: item.Price,
      })),
      total_price: apiData.TotalPrice,
      status: apiData.Status,
      created_at: apiData.CreatedAt,
      updated_at: apiData.UpdatedAt,
    };
    
    return transformedOrder;
  }

  async getOrder(id: string): Promise<RoomServiceOrder> {
    const response = await this.get<any>(`/orders/${id}`);
    
    // Transform API response (PascalCase) to TypeScript types (snake_case)
    const apiData = response.data;
    const transformedOrder: RoomServiceOrder = {
      id: apiData.ID,
      room_id: parseInt(apiData.RoomID),
      reservation_id: parseInt(apiData.ReservationID),
      billing_folder_id: apiData.BillingFolderID,
      billing_ticket_id: apiData.BillingTicketID || "",
      items: apiData.Items.map((item: any) => ({
        menu_item_id: item.MenuItemID,
        quantity: item.Quantity,
        unit_price: item.Price,
      })),
      total_price: apiData.TotalPrice,
      status: apiData.Status,
      created_at: apiData.CreatedAt,
      updated_at: apiData.UpdatedAt,
    };
    
    return transformedOrder;
  }

  async getOrdersByRoom(roomId: string): Promise<RoomServiceOrder[]> {
    const response = await this.get<any[]>(`/orders/room/${roomId}`);
    
    // Transform API response (PascalCase) to TypeScript types (snake_case)
    return response.data.map((apiData: any) => ({
      id: apiData.ID,
      room_id: parseInt(apiData.RoomID),
      reservation_id: parseInt(apiData.ReservationID),
      billing_folder_id: apiData.BillingFolderID,
      billing_ticket_id: apiData.BillingTicketID || "",
      items: apiData.Items.map((item: any) => ({
        menu_item_id: item.MenuItemID,
        quantity: item.Quantity,
        unit_price: item.Price,
      })),
      total_price: apiData.TotalPrice,
      status: apiData.Status,
      created_at: apiData.CreatedAt,
      updated_at: apiData.UpdatedAt,
    }));
  }
}

// ============================================================================
// Lobby Service Client
// Endpoints: /check_in/{room_number}, /check_out/{room_id}
// ============================================================================

export class LobbyClient extends HttpClient {
  constructor(options?: ServiceClientOptions) {
    super(config.services.lobby.url, options);
  }

  async checkIn(
    roomNumber: string,
    request?: CheckInRequest
  ): Promise<CheckInResponse> {
    const response = await this.post<CheckInResponse>(
      `/check_in/${roomNumber}`,
      request || {}
    );
    return response.data;
  }

  async checkOut(
    roomId: string,
    request?: CheckOutRequest
  ): Promise<CheckOutResponse> {
    const response = await this.post<CheckOutResponse>(
      `/check_out/${roomId}`,
      request || {}
    );
    return response.data;
  }
}

// ============================================================================
// Cleaning Service Client
// Endpoints: /jobs, /jobs/{id}, /staff, /damage-reports, /rooms/{id}/status
// Note: Requires authentication
// ============================================================================

export class CleaningClient extends HttpClient {
  constructor(options?: ServiceClientOptions) {
    // Add Basic Auth headers if credentials are provided in config
    const authOptions = { ...options };
    if (config.services.cleaning.username && config.services.cleaning.password) {
      const credentials = btoa(
        `${config.services.cleaning.username}:${config.services.cleaning.password}`
      );
      authOptions.headers = {
        ...authOptions.headers,
        Authorization: `Basic ${credentials}`,
      };
    }
    super(config.services.cleaning.url, authOptions);
  }

  // Cleaning Job operations
  async searchJobs(params?: SearchCleaningJobsParams): Promise<CleaningJob[]> {
    const queryParams = new URLSearchParams();
    if (params?.roomNumber) queryParams.append("roomNumber", params.roomNumber);
    if (params?.assignedTo !== undefined)
      queryParams.append("assignedTo", params.assignedTo.toString());
    if (params?.page !== undefined)
      queryParams.append("page", params.page.toString());
    if (params?.pageSize !== undefined)
      queryParams.append("pageSize", params.pageSize.toString());

    const query = queryParams.toString() ? `?${queryParams.toString()}` : "";
    const response = await this.get<CleaningJob[]>(`/jobs${query}`);
    return response.data;
  }

  async getJob(id: number): Promise<CleaningJob> {
    const response = await this.get<CleaningJob>(`/jobs/${id}`);
    return response.data;
  }

  async assignJob(id: number, staffId: number): Promise<MessageResponse> {
    const response = await this.post<MessageResponse>(`/jobs/${id}/assign`, {
      staffId,
    });
    return response.data;
  }

  async startJob(id: number): Promise<MessageResponse> {
    const response = await this.post<MessageResponse>(
      `/jobs/${id}/started`,
      {}
    );
    return response.data;
  }

  async finishJob(id: number): Promise<MessageResponse> {
    const response = await this.post<MessageResponse>(
      `/jobs/${id}/finished`,
      {}
    );
    return response.data;
  }

  async createCleaningJobs(): Promise<MessageResponse> {
    const response = await this.post<MessageResponse>(
      "/jobs/create-cleaning-jobs",
      {}
    );
    return response.data;
  }

  // Staff operations
  async createStaff(data: CreateCleaningStaffRequest): Promise<CleaningStaff> {
    const response = await this.post<CleaningStaff>("/staff", data);
    return response.data;
  }

  async searchStaff(): Promise<CleaningStaff[]> {
    const response = await this.get<CleaningStaff[]>("/staff");
    return response.data;
  }

  // Damage Report operations
  async createDamageReport(
    data: CreateDamageReportRequest
  ): Promise<DamageReport> {
    const response = await this.post<DamageReport>("/damage-reports", data);
    return response.data;
  }

  async getDamageReports(): Promise<DamageReport[]> {
    const response = await this.get<DamageReport[]>("/damage-reports");
    return response.data;
  }

  async getDamageReport(id: number): Promise<DamageReport> {
    const response = await this.get<DamageReport>(`/damage-reports/${id}`);
    return response.data;
  }

  // Room status operations
  async getRoomStatus(roomId: string): Promise<RoomOccupancyStatus> {
    const response = await this.get<RoomOccupancyStatus>(
      `/rooms/${roomId}/status`
    );
    return response.data;
  }
}

// ============================================================================
// Client Factory
// ============================================================================

export interface ServiceClients {
  billing: BillingClient;
  reservations: ReservationsClient;
  kitchen: KitchenClient;
  roomservice: RoomServiceClient;
  lobby: LobbyClient;
  cleaning: CleaningClient;
}

export function createServiceClients(
  options?: ServiceClientOptions
): ServiceClients {
  return {
    billing: new BillingClient(options),
    reservations: new ReservationsClient(options),
    kitchen: new KitchenClient(options),
    roomservice: new RoomServiceClient(options),
    lobby: new LobbyClient(options),
    cleaning: new CleaningClient(options),
  };
}

export const clients = createServiceClients();
