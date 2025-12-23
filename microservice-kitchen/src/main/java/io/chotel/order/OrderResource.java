package io.chotel.order;

import io.chotel.order.dto.Order;
import io.chotel.order.dto.OrderCompletionRequest;
import io.quarkus.logging.Log;
import io.smallrye.common.annotation.RunOnVirtualThread;

import jakarta.inject.Inject;
import jakarta.validation.Valid;
import jakarta.ws.rs.Consumes;
import jakarta.ws.rs.GET;
import jakarta.ws.rs.NotFoundException;
import jakarta.ws.rs.PATCH;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.PathParam;
import jakarta.ws.rs.Produces;
import jakarta.ws.rs.core.MediaType;

import java.util.List;

@Path("/orders")
public class OrderResource {

    @Inject
    OrderService orderService;

    @GET
    @Produces(MediaType.APPLICATION_JSON)
    public List<Order> getOrders() {
        Log.info("Getting orders");
        return orderService.getOrders();
    }

    @GET
    @Path("/{id}")
    @Produces(MediaType.APPLICATION_JSON)
    @RunOnVirtualThread
    public Order getOrder(@PathParam("id") String id) {
        Log.info("Getting order with id " + id);
        return orderService.getOrder(id).orElseThrow(NotFoundException::new);
    }

    @PATCH
    @Path("/{id}")
    @Consumes(MediaType.APPLICATION_JSON)
    @RunOnVirtualThread
    public void updateOrder(@PathParam("id") String id, @Valid OrderCompletionRequest request) {
        Log.info("Updating order with id " + id + " to " + request.getStatus());
        orderService.updateOrder(id, request.getStatus());
    }
}
