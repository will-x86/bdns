use crate::proto::proto::category_service_client::CategoryServiceClient;
use crate::proto::proto::{
    BlockCategoryRequest, CategoryBlock, ListBlockedRequest, UnblockCategoryRequest,
};
use crate::router::AppError;
use tonic::transport::Channel;

#[derive(Clone)]
pub struct CategorySvc {
    client: CategoryServiceClient<Channel>,
}

impl CategorySvc {
    pub async fn new(addr: String) -> Result<Self, AppError> {
        let client = CategoryServiceClient::connect(addr).await?;
        Ok(Self { client })
    }

    pub async fn list_blocked(self, profile_id: String) -> Result<Vec<String>, AppError> {
        let request = ListBlockedRequest { profile_id };
        let response = self.client.clone().list_blocked(request).await?;
        Ok(response.into_inner().categories)
    }

    pub async fn block_category(
        self,
        profile_id: String,
        category: String,
    ) -> Result<CategoryBlock, AppError> {
        let request = BlockCategoryRequest {
            profile_id,
            category,
        };
        let response = self.client.clone().block_category(request).await?;
        Ok(response.into_inner())
    }

    pub async fn unblock_category(
        self,
        profile_id: String,
        category: String,
    ) -> Result<(), AppError> {
        let request = UnblockCategoryRequest {
            profile_id,
            category,
        };
        let _response = self.client.clone().unblock_category(request).await?;
        Ok(())
    }
}
