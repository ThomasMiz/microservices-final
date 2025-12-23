package io.chotel.inventory;

import io.chotel.inventory.dto.InventoryBulkUpdateRequest;
import io.chotel.inventory.dto.InventoryResponse;
import io.chotel.inventory.dto.InventoryUpdateRequest;
import io.quarkus.logging.Log;
import io.smallrye.common.annotation.RunOnVirtualThread;

import jakarta.inject.Inject;
import jakarta.validation.Valid;
import jakarta.ws.rs.Consumes;
import jakarta.ws.rs.GET;
import jakarta.ws.rs.POST;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.PathParam;
import jakarta.ws.rs.Produces;
import jakarta.ws.rs.core.MediaType;

import org.jboss.resteasy.reactive.RestResponse;

import java.util.List;

@Path("/inventory")
@Produces(MediaType.APPLICATION_JSON)
@Consumes(MediaType.APPLICATION_JSON)
public class InventoryResource {

    @Inject InventoryService inventoryService;

    @POST
    @RunOnVirtualThread
    public RestResponse<InventoryResponse> updateStock(@Valid InventoryUpdateRequest request) {
        Log.infof("REST: Updating stock for %s to %d", request.menuItemId(), request.quantity());

        InventoryResponse response = inventoryService.updateStock(request);

        return RestResponse.ok(response);
    }

    @GET
    @Path("/{menuItemId}")
    @RunOnVirtualThread
    public RestResponse<InventoryResponse> getStock(@PathParam("menuItemId") String menuItemId) {
        Log.infof("REST: Getting stock for %s", menuItemId);

        InventoryResponse response = inventoryService.getStock(menuItemId);

        return RestResponse.ok(response);
    }

    @GET
    @RunOnVirtualThread
    public RestResponse<List<InventoryResponse>> listAllInventory() {
        Log.info("REST: Listing all inventory items");

        List<InventoryResponse> inventory = inventoryService.listAllInventory();

        return RestResponse.ok(inventory);
    }

    @POST
    @Path("/bulk")
    @RunOnVirtualThread
    public RestResponse<List<InventoryResponse>> bulkUpdateStock(
            @Valid InventoryBulkUpdateRequest request) {
        Log.infof("REST: Bulk updating %d inventory items", request.items().size());

        List<InventoryResponse> responses = inventoryService.bulkUpdateStock(request);

        return RestResponse.ok(responses);
    }

    @POST
    @Path("/{menuItemId}/add")
    @RunOnVirtualThread
    public RestResponse<?> addStock(
            @PathParam("menuItemId") String menuItemId, @Valid InventoryUpdateRequest request) {
        Log.infof("REST: Adding %d units to stock for %s", request.quantity(), menuItemId);

        if (!menuItemId.equals(request.menuItemId())) {
            return RestResponse.ResponseBuilder.create(RestResponse.Status.BAD_REQUEST)
                .entity("Menu item ID in path must match request body")
                .build();
        }

        InventoryResponse response = inventoryService.addStock(request);

        return RestResponse.ok(response);
    }
}
