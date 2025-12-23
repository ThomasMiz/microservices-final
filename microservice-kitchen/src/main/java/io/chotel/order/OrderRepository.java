
package io.chotel.order;

import java.util.List;
import java.util.Optional;

import io.chotel.order.dto.Order;
import io.opentelemetry.instrumentation.annotations.WithSpan;
import io.quarkus.redis.datasource.RedisDataSource;
import io.quarkus.redis.datasource.hash.HashCommands;
import jakarta.enterprise.context.ApplicationScoped;

@ApplicationScoped
public class OrderRepository {

    private static final String ORDER_KEY = "orders";

    private final HashCommands<String, String, Order> commands;

    public OrderRepository(RedisDataSource redisDataSource) {
        commands = redisDataSource.hash(Order.class);
    }

    @WithSpan("order.getOrder")
    public Optional<Order> getOrder(String id) {
        return Optional.ofNullable(commands.hget(ORDER_KEY, id));
    }

    @WithSpan("order.setOrder")
    public void createOrder(Order order) {
        commands.hset(ORDER_KEY, order.id(), order);
    }

    @WithSpan("order.deleteOrder")
    public void deleteOrder(String id) {
        commands.hdel(ORDER_KEY, id);
    }

    @WithSpan("order.getOrders")
    public List<Order> getOrders() {
        return commands.hvals(ORDER_KEY);
    }
}
