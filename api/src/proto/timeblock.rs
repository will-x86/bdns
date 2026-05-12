use crate::proto::proto::time_block_service_client::TimeBlockServiceClient;
use crate::proto::proto::{
    CreateTimeBlockRequest, DeleteTimeBlockRequest, ListTimeBlocksRequest, TimeBlock,
    TimeBlockList,
};
use crate::router::AppError;
use tonic::transport::Channel;

#[derive(Clone)]
pub struct TimeBlockSvc {
    client: TimeBlockServiceClient<Channel>,
}

impl TimeBlockSvc {
    pub async fn new(addr: String) -> Result<Self, AppError> {
        let client = TimeBlockServiceClient::connect(addr).await?;
        Ok(Self { client })
    }

    pub async fn list(&self, req: ListTimeBlocksRequest) -> Result<TimeBlockList, AppError> {
        let response = self.client.clone().list(req).await?;
        Ok(response.into_inner())
    }

    pub async fn create(&self, req: CreateTimeBlockRequest) -> Result<TimeBlock, AppError> {
        let response = self.client.clone().create(req).await?;
        Ok(response.into_inner())
    }

    pub async fn delete(&self, req: DeleteTimeBlockRequest) -> Result<(), AppError> {
        self.client.clone().delete(req).await?;
        Ok(())
    }
}
