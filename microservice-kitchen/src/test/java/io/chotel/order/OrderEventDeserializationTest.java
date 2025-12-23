package io.chotel.order;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import io.chotel.order.dto.Item;
import io.chotel.order.dto.OrderEvent;
import io.quarkus.test.junit.QuarkusTest;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.time.Instant;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Tests for OrderEvent JSON deserialization.
 * This test validates that the critical fixes are working:
 * 1. Instant timestamp handling (instead of LocalDateTime) to support ISO-8601 with 'Z' suffix
 * 2. snake_case field mapping for menu_item_id in Item records
 */
@QuarkusTest
class OrderEventDeserializationTest {

    private ObjectMapper objectMapper;

    @BeforeEach
    void setUp() {
        objectMapper = new ObjectMapper();
        objectMapper.registerModule(new JavaTimeModule());
    }

    /**
     * Test Case 1: Verify that OrderEvent can deserialize timestamps with 'Z' suffix (UTC)
     * This was the primary issue - LocalDateTime cannot handle timezone information.
     */
    @Test
    void testDeserializeOrderEventWithZuluTimestamp() throws Exception {
        String json = """
            {
                "event_type": "OrderRequested",
                "order_id": "order-123",
                "room_id": "room-456",
                "reservation_id": "reservation-789",
                "items": [
                    {
                        "menu_item_id": "burger-deluxe",
                        "quantity": 2,
                        "price": "15.99"
                    }
                ],
                "total_price": "31.98",
                "billing_folder_id": null,
                "reason": null,
                "timestamp": "2025-12-21T20:50:15.123456Z"
            }
            """;

        // This should NOT throw InvalidFormatException anymore
        OrderEvent event = objectMapper.readValue(json, OrderEvent.class);

        assertNotNull(event);
        assertEquals("OrderRequested", event.eventType());
        assertEquals("order-123", event.orderId());
        assertEquals("room-456", event.roomId());
        assertEquals("reservation-789", event.reservationId());
        assertEquals("31.98", event.totalPrice());
        
        // Verify timestamp is correctly parsed as Instant
        assertNotNull(event.timestamp());
        assertEquals(Instant.parse("2025-12-21T20:50:15.123456Z"), event.timestamp());
    }

    /**
     * Test Case 2: Verify that Item correctly deserializes snake_case menu_item_id
     * Without @JsonProperty("menu_item_id"), Jackson would look for "menuItemId" and fail.
     */
    @Test
    void testDeserializeItemWithSnakeCaseField() throws Exception {
        String json = """
            {
                "event_type": "OrderRequested",
                "order_id": "order-123",
                "room_id": "room-456",
                "reservation_id": "reservation-789",
                "items": [
                    {
                        "menu_item_id": "burger-deluxe",
                        "quantity": 2,
                        "price": "15.99"
                    },
                    {
                        "menu_item_id": "fries-large",
                        "quantity": 1,
                        "price": "5.99"
                    }
                ],
                "total_price": "37.97",
                "billing_folder_id": null,
                "reason": null,
                "timestamp": "2025-12-21T20:50:15Z"
            }
            """;

        OrderEvent event = objectMapper.readValue(json, OrderEvent.class);

        assertNotNull(event);
        assertNotNull(event.items());
        assertEquals(2, event.items().size());
        
        Item firstItem = event.items().get(0);
        assertNotNull(firstItem.menuItemId(), "menuItemId should not be null - check @JsonProperty annotation");
        assertEquals("burger-deluxe", firstItem.menuItemId());
        assertEquals(2, firstItem.quantity());
        assertEquals("15.99", firstItem.price());
        
        Item secondItem = event.items().get(1);
        assertNotNull(secondItem.menuItemId(), "menuItemId should not be null - check @JsonProperty annotation");
        assertEquals("fries-large", secondItem.menuItemId());
        assertEquals(1, secondItem.quantity());
        assertEquals("5.99", secondItem.price());
    }

    /**
     * Test Case 3: Verify all OrderEvent event types can be deserialized
     */
    @Test
    void testDeserializeOrderAcceptedEvent() throws Exception {
        String json = """
            {
                "event_type": "OrderAccepted",
                "order_id": "order-123",
                "room_id": "room-456",
                "reservation_id": "reservation-789",
                "items": [
                    {
                        "menu_item_id": "burger-deluxe",
                        "quantity": 2,
                        "price": "15.99"
                    }
                ],
                "total_price": "31.98",
                "billing_folder_id": null,
                "reason": null,
                "timestamp": "2025-12-21T20:50:15Z"
            }
            """;

        OrderEvent event = objectMapper.readValue(json, OrderEvent.class);
        
        assertNotNull(event);
        assertEquals("OrderAccepted", event.eventType());
        assertFalse(event.isOrderRequested());
        assertFalse(event.isOrderConfirmed());
        assertFalse(event.isOrderBillingFailed());
    }

    @Test
    void testDeserializeOrderRejectedEvent() throws Exception {
        String json = """
            {
                "event_type": "OrderRejected",
                "order_id": "order-123",
                "room_id": "room-456",
                "reservation_id": "reservation-789",
                "items": [
                    {
                        "menu_item_id": "burger-deluxe",
                        "quantity": 2,
                        "price": "15.99"
                    }
                ],
                "total_price": "31.98",
                "billing_folder_id": null,
                "reason": "Insufficient stock for item: burger-deluxe",
                "timestamp": "2025-12-21T20:50:15Z"
            }
            """;

        OrderEvent event = objectMapper.readValue(json, OrderEvent.class);
        
        assertNotNull(event);
        assertEquals("OrderRejected", event.eventType());
        assertEquals("Insufficient stock for item: burger-deluxe", event.reason());
    }

    @Test
    void testDeserializeOrderConfirmedEvent() throws Exception {
        String json = """
            {
                "event_type": "OrderConfirmed",
                "order_id": "order-123",
                "room_id": "room-456",
                "reservation_id": "reservation-789",
                "items": [
                    {
                        "menu_item_id": "burger-deluxe",
                        "quantity": 2,
                        "price": "15.99"
                    }
                ],
                "total_price": "31.98",
                "billing_folder_id": "billing-folder-456",
                "reason": null,
                "timestamp": "2025-12-21T20:50:15Z"
            }
            """;

        OrderEvent event = objectMapper.readValue(json, OrderEvent.class);
        
        assertNotNull(event);
        assertTrue(event.isOrderConfirmed());
        assertEquals("billing-folder-456", event.billingFolderId());
    }

    @Test
    void testDeserializeOrderBillingFailedEvent() throws Exception {
        String json = """
            {
                "event_type": "OrderBillingFailed",
                "order_id": "order-123",
                "room_id": "room-456",
                "reservation_id": "reservation-789",
                "items": [
                    {
                        "menu_item_id": "burger-deluxe",
                        "quantity": 2,
                        "price": "15.99"
                    }
                ],
                "total_price": "31.98",
                "billing_folder_id": "billing-folder-456",
                "reason": "Payment declined",
                "timestamp": "2025-12-21T20:50:15Z"
            }
            """;

        OrderEvent event = objectMapper.readValue(json, OrderEvent.class);
        
        assertNotNull(event);
        assertTrue(event.isOrderBillingFailed());
        assertEquals("Payment declined", event.reason());
    }

    /**
     * Test Case 4: Verify that timestamp with milliseconds is handled correctly
     */
    @Test
    void testDeserializeTimestampWithMilliseconds() throws Exception {
        String json = """
            {
                "event_type": "OrderRequested",
                "order_id": "order-123",
                "room_id": "room-456",
                "reservation_id": "reservation-789",
                "items": [],
                "total_price": "0.00",
                "billing_folder_id": null,
                "reason": null,
                "timestamp": "2025-12-21T20:50:15.999Z"
            }
            """;

        OrderEvent event = objectMapper.readValue(json, OrderEvent.class);
        
        assertNotNull(event.timestamp());
        assertEquals(Instant.parse("2025-12-21T20:50:15.999Z"), event.timestamp());
    }

    /**
     * Test Case 5: Verify that timestamp without fractional seconds is handled correctly
     */
    @Test
    void testDeserializeTimestampWithoutFractionalSeconds() throws Exception {
        String json = """
            {
                "event_type": "OrderRequested",
                "order_id": "order-123",
                "room_id": "room-456",
                "reservation_id": "reservation-789",
                "items": [],
                "total_price": "0.00",
                "billing_folder_id": null,
                "reason": null,
                "timestamp": "2025-12-21T20:50:15Z"
            }
            """;

        OrderEvent event = objectMapper.readValue(json, OrderEvent.class);
        
        assertNotNull(event.timestamp());
        assertEquals(Instant.parse("2025-12-21T20:50:15Z"), event.timestamp());
    }

    /**
     * Test Case 6: Verify empty items list is handled correctly
     */
    @Test
    void testDeserializeWithEmptyItems() throws Exception {
        String json = """
            {
                "event_type": "OrderRequested",
                "order_id": "order-123",
                "room_id": "room-456",
                "reservation_id": "reservation-789",
                "items": [],
                "total_price": "0.00",
                "billing_folder_id": null,
                "reason": null,
                "timestamp": "2025-12-21T20:50:15Z"
            }
            """;

        OrderEvent event = objectMapper.readValue(json, OrderEvent.class);
        
        assertNotNull(event);
        assertNotNull(event.items());
        assertTrue(event.items().isEmpty());
    }
}
