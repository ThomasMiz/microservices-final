package io.chotel.cleaning.external.billing;

import io.chotel.cleaning.external.billing.model.BillingFolder;
import io.chotel.cleaning.external.billing.model.BillingTicket;
import io.chotel.cleaning.external.billing.model.CreateTicketRequest;

public interface BillingService {

    BillingFolder createBillingFolder(String name);

    void deleteBillingFolder(String folderId);

    BillingTicket createTicket(CreateTicketRequest request);

    void deleteTicket(String ticketId);
}
