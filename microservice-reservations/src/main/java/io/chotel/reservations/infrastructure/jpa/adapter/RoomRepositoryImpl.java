package io.chotel.reservations.infrastructure.jpa.adapter;

import io.chotel.reservations.domain.model.Room;
import io.chotel.reservations.domain.model.request.CreateRoomRequest;
import io.chotel.reservations.domain.model.request.SearchRoomsRequest;
import io.chotel.reservations.domain.model.request.SortDirection;
import io.chotel.reservations.domain.model.result.PageResult;
import io.chotel.reservations.domain.repository.RoomRepository;
import io.chotel.reservations.infrastructure.jpa.entity.ReservationJpaEntity;
import io.chotel.reservations.infrastructure.jpa.entity.RoomJpaEntity;
import io.chotel.reservations.infrastructure.jpa.mapper.RoomJpaMapper;
import io.chotel.reservations.infrastructure.jpa.repository.RoomJpaRepository;
import jakarta.inject.Singleton;
import jakarta.persistence.*;
import jakarta.persistence.criteria.Expression;
import jakarta.persistence.criteria.Order;
import jakarta.persistence.criteria.ParameterExpression;
import lombok.RequiredArgsConstructor;
import org.apache.commons.lang3.StringUtils;
import org.hibernate.query.criteria.HibernateCriteriaBuilder;
import org.hibernate.query.criteria.JpaCriteriaQuery;
import org.hibernate.query.criteria.JpaRoot;
import org.hibernate.query.criteria.JpaSubQuery;

import java.math.BigDecimal;
import java.time.OffsetDateTime;
import java.util.*;
import java.util.function.Consumer;
import java.util.function.Function;
import java.util.function.Supplier;

import static io.chotel.reservations.domain.util.ObjectUtils.coalesce;

@Singleton
@RequiredArgsConstructor
public class RoomRepositoryImpl implements RoomRepository {

    @PersistenceContext
    private final EntityManager entityManager;

    private final RoomJpaRepository roomJpaRepository;

    @Override
    public Optional<Room> findById(Long id) {
        return roomJpaRepository.findById(id).map(RoomJpaMapper::toDomain);
    }

    @Override
    public Room createRoom(CreateRoomRequest request) {
        RoomJpaEntity entity = new RoomJpaEntity(
                request.number(),
                request.active(),
                request.name(),
                request.description(),
                request.maxCapacity(),
                request.category(),
                request.hourlyPrice()
        );

        entity = roomJpaRepository.save(entity);
        return RoomJpaMapper.toDomain(entity);
    }

    @Override
    public PageResult<Room> search(SearchRoomsRequest request) {
        PageResult<RoomJpaEntity> page = searchInternal(request);
        return page.map(RoomJpaMapper::toDomain);
    }

    private PageResult<RoomJpaEntity> searchInternal(SearchRoomsRequest request) {
        HibernateCriteriaBuilder cb = (HibernateCriteriaBuilder) entityManager.getCriteriaBuilder();
        JpaCriteriaQuery<RoomJpaEntity> cq = cb.createQuery(RoomJpaEntity.class);

        JpaRoot<RoomJpaEntity> room = cq.from(RoomJpaEntity.class);

        List<Expression<Boolean>> whereClauses = new ArrayList<>();

        String textSearchStr = StringUtils.trimToNull(StringUtils.stripAccents(request.textSearch()));
        ParameterExpression<String> textSearch = textSearchStr == null ? null : cb.parameter(String.class, "textSearch");
        ParameterExpression<Boolean> active = request.active() == null ? null : cb.parameter(Boolean.class, "active");
        ParameterExpression<OffsetDateTime> availableAfter = request.availableAfter() == null ? null : cb.parameter(OffsetDateTime.class, "availableAfter");
        ParameterExpression<OffsetDateTime> availableBefore = request.availableBefore() == null ? null : cb.parameter(OffsetDateTime.class, "availableBefore");
        ParameterExpression<BigDecimal> minPrice = request.minPrice() == null ? null : cb.parameter(BigDecimal.class, "minPrice");
        ParameterExpression<BigDecimal> maxPrice = request.maxPrice() == null ? null : cb.parameter(BigDecimal.class, "maxPrice");
        ParameterExpression<Integer> minCapacity = request.minCapacity() == null ? null : cb.parameter(Integer.class, "minCapacity");
        ParameterExpression<Integer> maxCapacity = request.maxCapacity() == null ? null : cb.parameter(Integer.class, "maxCapacity");
        ParameterExpression<Set> categories = request.categories() == null || request.categories().isEmpty() ? null : cb.parameter(Set.class, "categories");


        Expression<Double> searchSimilarity = null;
        if (textSearch != null) {
            Expression<String> searchIndexedExpr = cb.concat(
                    cb.coalesce(room.get("name"), cb.literal("")),
                    cb.coalesce(room.get("description"), cb.literal(""))
            );

            searchSimilarity = cb.function("word_similarity", Double.class, textSearch, searchIndexedExpr);
            whereClauses.add(cb.greaterThanOrEqualTo(searchSimilarity, cb.literal(0.5)));
        }

        if (active != null) {
            whereClauses.add(cb.equal(room.get("active"), active));
        }

        if (availableAfter != null) {
            JpaSubQuery<Integer> sq = cq.subquery(Integer.class);
            sq.select(cb.literal(1));
            JpaRoot<ReservationJpaEntity> reservation = sq.from(ReservationJpaEntity.class);
            List<Expression<Boolean>> subqueryWhereclauses = new ArrayList<>();

            subqueryWhereclauses.add(cb.equal(reservation.get("room"), room));

            subqueryWhereclauses.add(cb.greaterThan(reservation.get("endDate"), availableAfter));
            if (availableBefore != null) {
                subqueryWhereclauses.add(cb.lessThan(reservation.get("startDate"), availableBefore));
            }

            subqueryWhereclauses.stream().reduce(cb::and).ifPresent(sq::where);
            whereClauses.add(cb.not(cb.exists(sq)));
        }

        if (minPrice != null) {
            whereClauses.add(cb.greaterThanOrEqualTo(room.get("hourlyPrice"), minPrice));
        }

        if (maxPrice != null) {
            whereClauses.add(cb.lessThanOrEqualTo(room.get("hourlyPrice"), maxPrice));
        }

        if (minCapacity != null) {
            whereClauses.add(cb.greaterThanOrEqualTo(room.get("maxCapacity"), minCapacity));
        }

        if (maxCapacity != null) {
            whereClauses.add(cb.lessThanOrEqualTo(room.get("maxCapacity"), maxCapacity));
        }

        if (categories != null) {
            whereClauses.add(room.get("category").in(categories));
        }

        whereClauses.stream().reduce(cb::and).ifPresent(cq::where);

        Consumer<Query> bindQueryParams = query -> {
            Map<String, Supplier<Object>> paramValueGetter = Map.of(
                    "textSearch", request::textSearch,
                    "active", request::active,
                    "availableAfter", request::availableAfter,
                    "availableBefore", request::availableBefore,
                    "minPrice", request::minPrice,
                    "maxPrice", request::maxPrice,
                    "minCapacity", request::minCapacity,
                    "maxCapacity", request::maxCapacity,
                    "categories", request::categories
            );

            for (Parameter<?> parameter : query.getParameters()) {
                query.setParameter(parameter.getName(), paramValueGetter.get(parameter.getName()).get());
            }
        };

        TypedQuery<Long> countQuery = entityManager.createQuery(cq.createCountQuery());
        bindQueryParams.accept(countQuery);
        long totalElements = countQuery.getSingleResult();

        SearchRoomsRequest.SortBy sortBy = request.pageRequest().sortBy();
        if (sortBy == null || (searchSimilarity == null && sortBy == SearchRoomsRequest.SortBy.SEARCH_SIMILARITY)) {
            sortBy = SearchRoomsRequest.SortBy.ID;
        }

        List<Expression<?>> sortExpressions = switch (sortBy) {
            case ID -> List.of(room.get("id"));
            case NUMBER -> List.of(room.get("number"), room.get("id"));
            case PRICE -> List.of(room.get("hourlyPrice"), room.get("id"));
            case SEARCH_SIMILARITY -> List.of(searchSimilarity, room.get("id"));
        };

        SortDirection sortDirection = coalesce(request.pageRequest().sortDirection(), SortDirection.ASCENDING);
        Function<Expression<?>, Order> sorter = sortDirection == SortDirection.ASCENDING ? cb::asc : cb::desc;
        cq.orderBy(sortExpressions.stream().map(sorter).toList());

        TypedQuery<RoomJpaEntity> typedQuery = entityManager.createQuery(cq);
        bindQueryParams.accept(typedQuery);
        typedQuery.setFirstResult(request.pageRequest().firstElementOffset());
        typedQuery.setMaxResults(request.pageRequest().pageSize());
        List<RoomJpaEntity> contents = typedQuery.getResultList();

        return new PageResult<>(contents, totalElements);
    }
}
