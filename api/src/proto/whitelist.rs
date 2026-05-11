use crate::proto::proto::whitelist_service_client::WhitelistServiceClient;
use crate::proto::proto::{
    AddPermanentRequest, AddTemporaryRequest, ListPermanentRequest, ListTemporaryRequest,
    RemovePermanentRequest, RemoveTemporaryRequest, WhitelistDomain, WhitelistDomainTemp,
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

    pub async fn list_permanent(self, profile_id: String) -> Result<Vec<String>, AppError> {
        let request = ListPermanentRequest { profile_id };
        let response = self.client.clone().list_permanent(request).await?;
        Ok(response.into_inner().domains)
    }

    pub async fn add_permanent(
        self,
        profile_id: String,
        domain: String,
    ) -> Result<WhitelistDomain, AppError> {
        let request = AddPermanentRequest { profile_id, domain };
        let response = self.client.clone().add_permanent(request).await?;
        Ok(response.into_inner())
    }

    pub async fn remove_permanent(
        self,
        profile_id: String,
        domain: String,
    ) -> Result<(), AppError> {
        let request = RemovePermanentRequest { profile_id, domain };
        let _response = self.client.clone().remove_permanent(request).await?;
        Ok(())
    }

    pub async fn list_temporary(
        self,
        profile_id: String,
    ) -> Result<Vec<WhitelistDomainTemp>, AppError> {
        let request = ListTemporaryRequest { profile_id };
        let response = self.client.clone().list_temporary(request).await?;
        Ok(response.into_inner().entries)
    }

    pub async fn add_temporary(
        self,
        profile_id: String,
        domain: String,
        expires_at: i64,
    ) -> Result<WhitelistDomainTemp, AppError> {
        let request = AddTemporaryRequest {
            profile_id,
            domain,
            expires_at,
        };
        let response = self.client.clone().add_temporary(request).await?;
        Ok(response.into_inner())
    }

    pub async fn remove_temporary(
        self,
        profile_id: String,
        domain: String,
    ) -> Result<(), AppError> {
        let request = RemoveTemporaryRequest { profile_id, domain };
        let _response = self.client.clone().remove_temporary(request).await?;
        Ok(())
    }
}
