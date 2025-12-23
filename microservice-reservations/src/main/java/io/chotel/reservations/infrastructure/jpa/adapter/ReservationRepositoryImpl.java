package io.chotel.reservations.infrastructure.jpa.adapter;

import io.chotel.reservations.domain.model.Reservation;
import io.chotel.reservations.domain.model.request.SearchReservationsRequest;
import io.chotel.reservations.domain.model.request.SortDirection;
import io.chotel.reservations.domain.model.result.PageResult;
import io.chotel.reservations.domain.repository.ReservationRepository;
import io.chotel.reservations.infrastructure.jpa.entity.ReservationJpaEntity;
import io.chotel.reservations.infrastructure.jpa.entity.RoomJpaEntity;
import io.chotel.reservations.infrastructure.jpa.mapper.ReservationJpaMapper;
import io.chotel.reservations.infrastructure.jpa.repository.ReservationJpaRepository;
import io.chotel.reservations.infrastructure.jpa.repository.RoomJpaRepository;
import jakarta.inject.Singleton;
import jakarta.persistence.*;
import jakarta.persistence.criteria.Expression;
import jakarta.persistence.criteria.Order;
import jakarta.persistence.criteria.ParameterExpression;
import jakarta.persistence.criteria.Root;
import lombok.RequiredArgsConstructor;
import org.apache.commons.lang3.StringUtils;
import org.hibernate.query.criteria.HibernateCriteriaBuilder;
import org.hibernate.query.criteria.JpaCriteriaQuery;

import java.math.BigDecimal;
import java.time.OffsetDateTime;
import java.util.*;
import java.util.function.Consumer;
import java.util.function.Function;
import java.util.function.Supplier;
import java.util.stream.Collectors;

import static io.chotel.reservations.domain.util.ObjectUtils.coalesce;

@Singleton
@RequiredArgsConstructor
public class ReservationRepositoryImpl implements ReservationRepository {

    @PersistenceContext
    private final EntityManager entityManager;

    private final RoomJpaRepository roomJpaRepository;
    private final ReservationJpaRepository reservationJpaRepository;

    @Override
    public Optional<Reservation> findById(long id) {
        return reservationJpaRepository.findById(id).map(ReservationJpaMapper::toDomain);
    }

    @Override
    public Reservation createReservation(
            Long roomId,
            String guestId,
            OffsetDateTime startDate,
            OffsetDateTime endDate,
            BigDecimal rentedHourlyPrice,
            BigDecimal totalPrice
    ) {
        RoomJpaEntity room = roomJpaRepository.findById(roomId).orElseThrow();
        ReservationJpaEntity entity = new ReservationJpaEntity(room, guestId, startDate, endDate, null, null, rentedHourlyPrice, totalPrice);
        entity = reservationJpaRepository.save(entity);
        return ReservationJpaMapper.toDomain(entity);
    }

    @Override
    public Reservation setBillingInfo(
            Reservation reservation,
            String billingFolderId,
            String reservationBillingTicket
    ) {
        ReservationJpaEntity entity = ReservationJpaMapper.toEntity(reservation);
        // entity = reservationJpaRepository.merge(entity);

        entity.setBillingFolderId(billingFolderId);
        entity.setReservationBillingTicket(reservationBillingTicket);
        entity = reservationJpaRepository.update(entity);

        return ReservationJpaMapper.toDomain(entity);
    }

    @Override
    public void deleteReservationById(long id) {
        reservationJpaRepository.deleteById(id);
    }

    @Override
    public Map<Long, List<Reservation>> findReservationsForRoomsBetweenDates(
            Collection<Long> roomIds,
            OffsetDateTime startDate,
            OffsetDateTime endDate
    ) {
        startDate = coalesce(startDate, OffsetDateTime.MIN);
        endDate = coalesce(endDate, OffsetDateTime.MAX);

        return reservationJpaRepository
                .findReservationsForRoomsBetweenDates(roomIds, startDate, endDate)
                .stream()
                .map(ReservationJpaMapper::toDomain)
                .collect(Collectors.groupingBy(r -> r.room().id()));
    }

    @Override
    public PageResult<Reservation> search(SearchReservationsRequest request) {
        PageResult<ReservationJpaEntity> page = searchInternal(request);
        return page.map(ReservationJpaMapper::toDomain);
    }

    private PageResult<ReservationJpaEntity> searchInternal(SearchReservationsRequest request) {
        HibernateCriteriaBuilder cb = (HibernateCriteriaBuilder) entityManager.getCriteriaBuilder();
        JpaCriteriaQuery<ReservationJpaEntity> cq = cb.createQuery(ReservationJpaEntity.class);

        Root<ReservationJpaEntity> reservation = cq.from(ReservationJpaEntity.class);

        List<Expression<Boolean>> whereClauses = new ArrayList<>();

        ParameterExpression<OffsetDateTime> from = request.from() == null ? null : cb.parameter(OffsetDateTime.class, "from");
        ParameterExpression<OffsetDateTime> to = request.to() == null ? null : cb.parameter(OffsetDateTime.class, "to");
        ParameterExpression<Set> roomIds = request.roomIds() == null || request.roomIds().isEmpty() ? null : cb.parameter(Set.class, "roomIds");
        ParameterExpression<Set> roomNumbers = request.roomNumbers() == null || request.roomNumbers().isEmpty() ? null : cb.parameter(Set.class, "roomNumbers");
        ParameterExpression<String> guestId = StringUtils.isBlank(request.guestId()) ? null : cb.parameter(String.class, "guestId");

        if (from != null) {
            whereClauses.add(cb.greaterThanOrEqualTo(reservation.get("endDate"), from));
        }

        if (to != null) {
            whereClauses.add(cb.lessThanOrEqualTo(reservation.get("startDate"), to));
        }

        if (roomIds != null) {
            whereClauses.add(reservation.get("room").get("id").in(roomIds));
        }

        if (roomNumbers != null) {
            whereClauses.add(reservation.get("room").get("number").in(roomNumbers));
        }

        if (guestId != null) {
            whereClauses.add(cb.equal(reservation.get("guestId"), guestId));
        }

        whereClauses.stream().reduce(cb::and).ifPresent(cq::where);

        Consumer<Query> bindQueryParams = query -> {
            Map<String, Supplier<Object>> paramValueGetter = Map.of(
                    "from", request::from,
                    "to", request::to,
                    "roomIds", request::roomIds,
                    "roomNumbers", request::roomNumbers,
                    "guestId", request::guestId
            );

            for (Parameter<?> parameter : query.getParameters()) {
                query.setParameter(parameter.getName(), paramValueGetter.get(parameter.getName()).get());
            }
        };

        TypedQuery<Long> countQuery = entityManager.createQuery(cq.createCountQuery());
        bindQueryParams.accept(countQuery);
        long totalElements = countQuery.getSingleResult();

        SearchReservationsRequest.SortBy sortBy = coalesce(request.pageRequest().sortBy(), SearchReservationsRequest.SortBy.ID);

        List<Expression<?>> sortExpressions = switch (sortBy) {
            case ID -> List.of(reservation.get("id"));
            case START_DATE -> List.of(reservation.get("startDate"), reservation.get("id"));
            case END_DATE -> List.of(reservation.get("endDate"), reservation.get("id"));
            case ROOM_ID -> List.of(reservation.get("room").get("id"), reservation.get("id"));
        };

        SortDirection sortDirection = coalesce(request.pageRequest().sortDirection(), SortDirection.ASCENDING);
        Function<Expression<?>, Order> sorter = sortDirection == SortDirection.ASCENDING ? cb::asc : cb::desc;
        cq.orderBy(sortExpressions.stream().map(sorter).toList());

        TypedQuery<ReservationJpaEntity> typedQuery = entityManager.createQuery(cq);
        bindQueryParams.accept(typedQuery);
        typedQuery.setFirstResult(request.pageRequest().firstElementOffset());
        typedQuery.setMaxResults(request.pageRequest().pageSize());
        List<ReservationJpaEntity> contents = typedQuery.getResultList();

        return new PageResult<>(contents, totalElements);
    }
}
