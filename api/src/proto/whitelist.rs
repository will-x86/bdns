use crate::proto::proto::whitelist_service_client::WhitelistServiceClient;
use crate::proto::proto::{
    AddPermanentRequest, AddTemporaryRequest, ListPermanentRequest, ListTemporaryRequest,
    RemovePermanentRequest, RemoveTemporaryRequest, WhitelistDomain, WhitelistDomainTemp,
    WhitelistDomains, WhitelistDomainsTemp,
};
use crate::router::AppError;
use tonic::transport::Channel;

#[derive(Clone)]
pub struct WhitelistSvc {
    client: WhitelistServiceClient<Channel>,
}

impl WhitelistSvc {
    pub async fn new(addr: String) -> Result<Self, AppError> {
        let client = WhitelistServiceClient::connect(addr).await?;
        Ok(Self { client })
    }

    pub async fn list_permanent(&self, req: ListPermanentRequest) -> Result<WhitelistDomains, AppError> {
        let response = self.client.clone().list_permanent(req).await?;
        Ok(response.into_inner())
    }

    pub async fn add_permanent(&self, req: AddPermanentRequest) -> Result<WhitelistDomain, AppError> {
        let response = self.client.clone().add_permanent(req).await?;
        Ok(response.into_inner())
    }

    pub async fn remove_permanent(&self, req: RemovePermanentRequest) -> Result<(), AppError> {
        self.client.clone().remove_permanent(req).await?;
        Ok(())
    }

    pub async fn list_temporary(&self, req: ListTemporaryRequest) -> Result<WhitelistDomainsTemp, AppError> {
        let response = self.client.clone().list_temporary(req).await?;
        Ok(response.into_inner())
    }

    pub async fn add_temporary(&self, req: AddTemporaryRequest) -> Result<WhitelistDomainTemp, AppError> {
        let response = self.client.clone().add_temporary(req).await?;
        Ok(response.into_inner())
    }

    pub async fn remove_temporary(&self, req: RemoveTemporaryRequest) -> Result<(), AppError> {
        self.client.clone().remove_temporary(req).await?;
        Ok(())
    }
}
