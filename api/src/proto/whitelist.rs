use crate::proto::proto::whitelist_service_client::WhitelistServiceClient;
use crate::proto::proto::{
    AddPermanentRequest, AddTemporaryRequest, ListPermanentRequest, ListTemporaryRequest,
    RemovePermanentRequest, RemoveTemporaryRequest, WhitelistDomain, WhitelistDomainTemp,
};
use tonic::transport::Channel;

pub struct WhitelistSvc {
    client: WhitelistServiceClient<Channel>,
}

impl WhitelistSvc {
    pub async fn new(addr: String) -> Result<Self, Box<dyn std::error::Error>> {
        let client = WhitelistServiceClient::connect(addr).await?;
        Ok(Self { client })
    }

    pub async fn list_permanent(
        &mut self,
        profile_id: String,
    ) -> Result<Vec<String>, Box<dyn std::error::Error>> {
        let request = ListPermanentRequest { profile_id };
        let response = self.client.list_permanent(request).await?;
        Ok(response.into_inner().domains)
    }

    pub async fn add_permanent(
        &mut self,
        profile_id: String,
        domain: String,
    ) -> Result<WhitelistDomain, Box<dyn std::error::Error>> {
        let request = AddPermanentRequest { profile_id, domain };
        let response = self.client.add_permanent(request).await?;
        Ok(response.into_inner())
    }

    pub async fn remove_permanent(
        &mut self,
        profile_id: String,
        domain: String,
    ) -> Result<(), Box<dyn std::error::Error>> {
        let request = RemovePermanentRequest { profile_id, domain };
        let _response = self.client.remove_permanent(request).await?;
        Ok(())
    }

    pub async fn list_temporary(
        &mut self,
        profile_id: String,
    ) -> Result<Vec<WhitelistDomainTemp>, Box<dyn std::error::Error>> {
        let request = ListTemporaryRequest { profile_id };
        let response = self.client.list_temporary(request).await?;
        Ok(response.into_inner().entries)
    }

    pub async fn add_temporary(
        &mut self,
        profile_id: String,
        domain: String,
        expires_at: i64,
    ) -> Result<WhitelistDomainTemp, Box<dyn std::error::Error>> {
        let request = AddTemporaryRequest {
            profile_id,
            domain,
            expires_at,
        };
        let response = self.client.add_temporary(request).await?;
        Ok(response.into_inner())
    }

    pub async fn remove_temporary(
        &mut self,
        profile_id: String,
        domain: String,
    ) -> Result<(), Box<dyn std::error::Error>> {
        let request = RemoveTemporaryRequest { profile_id, domain };
        let _response = self.client.remove_temporary(request).await?;
        Ok(())
    }
}

