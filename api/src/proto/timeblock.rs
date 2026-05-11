use crate::proto::proto::time_block_service_client::TimeBlockServiceClient;
use crate::proto::proto::{
    CreateTimeBlockRequest, DeleteTimeBlockRequest, ListTimeBlocksRequest, TimeBlock,
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

    pub async fn list(self, profile_id: String) -> Result<Vec<TimeBlock>, AppError> {
        let request = ListTimeBlocksRequest { profile_id };
        let response = self.client.clone().list(request).await?;
        Ok(response.into_inner().blocks)
    }

    pub async fn create(
        self,
        profile_id: String,
        category: String,
        start_time: i32,
        end_time: i32,
        day: i32,
    ) -> Result<TimeBlock, AppError> {
        let request = CreateTimeBlockRequest {
            profile_id,
            category,
            start_time,
            end_time,
            day,
        };
        let response = self.client.clone().create(request).await?;
        Ok(response.into_inner())
    }

    pub async fn delete(self, block_id: String) -> Result<(), AppError> {
        let request = DeleteTimeBlockRequest { block_id };
        let _response = self.client.clone().delete(request).await?;
        Ok(())
    }
}
