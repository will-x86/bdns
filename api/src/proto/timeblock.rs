use crate::proto::proto::time_block_service_client::TimeBlockServiceClient;
use crate::proto::proto::{
    CreateTimeBlockRequest, DeleteTimeBlockRequest, ListTimeBlocksRequest, TimeBlock,
};
use tonic::transport::Channel;

pub struct TimeBlockSvc {
    client: TimeBlockServiceClient<Channel>,
}

impl TimeBlockSvc {
    pub async fn new(addr: String) -> Result<Self, Box<dyn std::error::Error>> {
        let client = TimeBlockServiceClient::connect(addr).await?;
        Ok(Self { client })
    }

    pub async fn list(
        &mut self,
        profile_id: String,
    ) -> Result<Vec<TimeBlock>, Box<dyn std::error::Error>> {
        let request = ListTimeBlocksRequest { profile_id };
        let response = self.client.list(request).await?;
        Ok(response.into_inner().blocks)
    }

    pub async fn create(
        &mut self,
        profile_id: String,
        category: String,
        start_time: i32,
        end_time: i32,
        day: i32,
    ) -> Result<TimeBlock, Box<dyn std::error::Error>> {
        let request = CreateTimeBlockRequest {
            profile_id,
            category,
            start_time,
            end_time,
            day,
        };
        let response = self.client.create(request).await?;
        Ok(response.into_inner())
    }

    pub async fn delete(&mut self, block_id: String) -> Result<(), Box<dyn std::error::Error>> {
        let request = DeleteTimeBlockRequest { block_id };
        let _response = self.client.delete(request).await?;
        Ok(())
    }
}

