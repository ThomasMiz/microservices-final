package io.chotel.order.dto;

import com.fasterxml.jackson.annotation.JsonAlias;
import com.fasterxml.jackson.annotation.JsonProperty;

public record Item(
    @JsonProperty("menuItemId")
    @JsonAlias({"menu_item_id"})
    String menuItemId,
    @JsonProperty("quantity") int quantity,
    @JsonProperty("price") String price
) {}
