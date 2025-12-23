package io.chotel.reservations.domain.external.billing;

import io.chotel.reservations.domain.external.billing.model.BillingFolder;
import io.chotel.reservations.domain.external.billing.model.BillingTicket;
import io.chotel.reservations.domain.external.billing.model.CreateTicketRequest;

public interface BillingService {

    BillingFolder createBillingFolder(String name);

    void deleteBillingFolder(String folderId);

    BillingTicket createTicket(CreateTicketRequest request);

    void deleteTicket(String ticketId);
}
