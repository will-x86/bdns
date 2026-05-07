use crate::proto::proto::category_service_client::CategoryServiceClient;
use crate::proto::proto::{
    BlockCategoryRequest, CategoryBlock, ListBlockedRequest, UnblockCategoryRequest,
};
use tonic::transport::Channel;

pub struct CategorySvc {
    client: CategoryServiceClient<Channel>,
}

impl CategorySvc {
    pub async fn new(addr: String) -> Result<Self, Box<dyn std::error::Error>> {
        let client = CategoryServiceClient::connect(addr).await?;
        Ok(Self { client })
    }

    pub async fn list_blocked(
        &mut self,
        profile_id: String,
    ) -> Result<Vec<String>, Box<dyn std::error::Error>> {
        let request = ListBlockedRequest { profile_id };
        let response = self.client.list_blocked(request).await?;
        Ok(response.into_inner().categories)
    }

    pub async fn block_category(
        &mut self,
        profile_id: String,
        category: String,
    ) -> Result<CategoryBlock, Box<dyn std::error::Error>> {
        let request = BlockCategoryRequest {
            profile_id,
            category,
        };
        let response = self.client.block_category(request).await?;
        Ok(response.into_inner())
    }

    pub async fn unblock_category(
        &mut self,
        profile_id: String,
        category: String,
    ) -> Result<(), Box<dyn std::error::Error>> {
        let request = UnblockCategoryRequest {
            profile_id,
            category,
        };
        let _response = self.client.unblock_category(request).await?;
        Ok(())
    }
}

