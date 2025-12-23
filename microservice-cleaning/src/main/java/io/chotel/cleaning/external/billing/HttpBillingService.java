package io.chotel.cleaning.external.billing;

import io.chotel.cleaning.external.billing.config.BillingApiConfig;
import io.chotel.cleaning.external.billing.json.request.CreateBillingFolderRequestJson;
import io.chotel.cleaning.external.billing.json.request.CreateBillingTicketRequestJson;
import io.chotel.cleaning.external.billing.json.response.BillingFolderJson;
import io.chotel.cleaning.external.billing.json.response.BillingTicketJson;
import io.chotel.cleaning.external.billing.model.BillingFolder;
import io.chotel.cleaning.external.billing.model.BillingTicket;
import io.chotel.cleaning.external.billing.model.CreateTicketRequest;
import io.micronaut.http.HttpRequest;
import io.micronaut.http.client.BlockingHttpClient;
import io.micronaut.http.client.HttpClient;
import jakarta.inject.Singleton;

@Singleton
public class HttpBillingService implements BillingService {
    private final BlockingHttpClient httpClient;

    private final String baseUrl;
    private final String foldersUrl;
    private final String ticketsUrl;

    public HttpBillingService(BillingApiConfig apiConfig, HttpClient httpClient) {
        this.httpClient = httpClient.toBlocking();

        baseUrl = apiConfig.url();
        foldersUrl = baseUrl + "/folders";
        ticketsUrl = baseUrl + "/tickets";
    }

    @Override
    public BillingFolder createBillingFolder(String name) {
        CreateBillingFolderRequestJson requestBody = new CreateBillingFolderRequestJson(name);
        HttpRequest<CreateBillingFolderRequestJson> request = HttpRequest.POST(foldersUrl, requestBody);
        BillingFolderJson responseBody = httpClient.retrieve(request, BillingFolderJson.class);
        return responseBody.toDomain();
    }

    @Override
    public void deleteBillingFolder(String folderId) {
        HttpRequest<?> request = HttpRequest.DELETE(foldersUrl + "/" + folderId);
        httpClient.retrieve(request);
    }

    @Override
    public BillingTicket createTicket(CreateTicketRequest createTicketRequest) {
        CreateBillingTicketRequestJson requestBody = CreateBillingTicketRequestJson.from(createTicketRequest);
        HttpRequest<CreateBillingTicketRequestJson> request = HttpRequest.POST(ticketsUrl, requestBody);
        BillingTicketJson responseBody = httpClient.retrieve(request, BillingTicketJson.class);
        return responseBody.toDomain();
    }

    @Override
    public void deleteTicket(String ticketId) {
        HttpRequest<?> request = HttpRequest.DELETE(ticketsUrl + "/" + ticketId);
        httpClient.retrieve(request);
    }
}
