use crate::proto::proto::category_service_client::CategoryServiceClient;
use crate::proto::proto::{
    BlockCategoryRequest, CategoryBlock, CategoryList, ListBlockedRequest, UnblockCategoryRequest,
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

    pub async fn list_blocked(&self, req: ListBlockedRequest) -> Result<CategoryList, AppError> {
        let response = self.client.clone().list_blocked(req).await?;
        Ok(response.into_inner())
    }

    pub async fn block_category(&self, req: BlockCategoryRequest) -> Result<CategoryBlock, AppError> {
        let response = self.client.clone().block_category(req).await?;
        Ok(response.into_inner())
    }

    pub async fn unblock_category(&self, req: UnblockCategoryRequest) -> Result<(), AppError> {
        self.client.clone().unblock_category(req).await?;
        Ok(())
    }
}
