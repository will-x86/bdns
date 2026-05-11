use crate::proto::proto::pool_service_client::PoolServiceClient;
use crate::proto::proto::{
    BlockPoolCategoryRequest, CreatePoolRequest, CreditsResponse, DeletePoolRequest, FriendPool,
    GetCreditsRequest, GetPoolRequest, JoinPoolRequest, LeavePoolRequest, ListMembersRequest,
    ListPoolBlocksRequest, ListPoolsRequest, PoolBlock, PoolMember, UnblockPoolCategoryRequest,
};
use crate::router::AppError;
use tonic::transport::Channel;

#[derive(Clone)]
pub struct PoolSvc {
    client: PoolServiceClient<Channel>,
}

impl PoolSvc {
    pub async fn new(addr: String) -> Result<Self, AppError> {
        let client = PoolServiceClient::connect(addr).await?;
        Ok(Self { client })
    }

    pub async fn list_pools(self, user_id: String) -> Result<Vec<FriendPool>, AppError> {
        let request = ListPoolsRequest { user_id };
        let response = self.client.clone().list_pools(request).await?;
        Ok(response.into_inner().pools)
    }

    pub async fn create_pool(
        self,
        user_id: String,
        name: String,
        pool_mode: String,
        total_limit: i64,
    ) -> Result<FriendPool, AppError> {
        let request = CreatePoolRequest {
            user_id,
            name,
            pool_mode,
            total_limit,
        };
        let response = self.client.clone().create_pool(request).await?;
        Ok(response.into_inner())
    }

    pub async fn get_pool(self, pool_id: String) -> Result<FriendPool, AppError> {
        let request = GetPoolRequest { pool_id };
        let response = self.client.clone().get_pool(request).await?;
        Ok(response.into_inner())
    }

    pub async fn delete_pool(self, pool_id: String, user_id: String) -> Result<(), AppError> {
        let request = DeletePoolRequest { pool_id, user_id };
        let _response = self.client.clone().delete_pool(request).await?;
        Ok(())
    }

    pub async fn join_pool(self, pool_id: String, profile_id: String) -> Result<(), AppError> {
        let request = JoinPoolRequest {
            pool_id,
            profile_id,
        };
        let _response = self.client.clone().join_pool(request).await?;
        Ok(())
    }

    pub async fn leave_pool(self, pool_id: String, profile_id: String) -> Result<(), AppError> {
        let request = LeavePoolRequest {
            pool_id,
            profile_id,
        };
        let _response = self.client.clone().leave_pool(request).await?;
        Ok(())
    }

    pub async fn list_members(self, pool_id: String) -> Result<Vec<PoolMember>, AppError> {
        let request = ListMembersRequest { pool_id };
        let response = self.client.clone().list_members(request).await?;
        Ok(response.into_inner().members)
    }

    pub async fn list_blocks(self, pool_id: String) -> Result<Vec<PoolBlock>, AppError> {
        let request = ListPoolBlocksRequest { pool_id };
        let response = self.client.clone().list_blocks(request).await?;
        Ok(response.into_inner().blocks)
    }

    pub async fn block_category(self, pool_id: String, category: String) -> Result<(), AppError> {
        let request = BlockPoolCategoryRequest { pool_id, category };
        let _response = self.client.clone().block_category(request).await?;
        Ok(())
    }

    pub async fn unblock_category(self, pool_id: String, category: String) -> Result<(), AppError> {
        let request = UnblockPoolCategoryRequest { pool_id, category };
        let _response = self.client.clone().unblock_category(request).await?;
        Ok(())
    }

    pub async fn get_credits(
        self,
        pool_id: String,
        profile_id: String,
    ) -> Result<CreditsResponse, AppError> {
        let request = GetCreditsRequest {
            pool_id,
            profile_id,
        };
        let response = self.client.clone().get_credits(request).await?;
        Ok(response.into_inner())
    }
}
