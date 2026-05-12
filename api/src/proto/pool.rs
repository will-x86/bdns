use crate::proto::proto::pool_service_client::PoolServiceClient;
use crate::proto::proto::{
    BlockPoolCategoryRequest, CreatePoolRequest, CreditsResponse, DeletePoolRequest, FriendPool,
    GetCreditsRequest, GetPoolRequest, JoinPoolRequest, LeavePoolRequest, ListMembersRequest,
    ListPoolBlocksRequest, ListPoolsRequest, PoolBlockList, PoolList, PoolMemberList,
    UnblockPoolCategoryRequest,
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

    pub async fn list_pools(&self, req: ListPoolsRequest) -> Result<PoolList, AppError> {
        let response = self.client.clone().list_pools(req).await?;
        Ok(response.into_inner())
    }

    pub async fn create_pool(&self, req: CreatePoolRequest) -> Result<FriendPool, AppError> {
        let response = self.client.clone().create_pool(req).await?;
        Ok(response.into_inner())
    }

    pub async fn get_pool(&self, req: GetPoolRequest) -> Result<FriendPool, AppError> {
        let response = self.client.clone().get_pool(req).await?;
        Ok(response.into_inner())
    }

    pub async fn delete_pool(&self, req: DeletePoolRequest) -> Result<(), AppError> {
        self.client.clone().delete_pool(req).await?;
        Ok(())
    }

    pub async fn join_pool(&self, req: JoinPoolRequest) -> Result<(), AppError> {
        self.client.clone().join_pool(req).await?;
        Ok(())
    }

    pub async fn leave_pool(&self, req: LeavePoolRequest) -> Result<(), AppError> {
        self.client.clone().leave_pool(req).await?;
        Ok(())
    }

    pub async fn list_members(&self, req: ListMembersRequest) -> Result<PoolMemberList, AppError> {
        let response = self.client.clone().list_members(req).await?;
        Ok(response.into_inner())
    }

    pub async fn list_blocks(&self, req: ListPoolBlocksRequest) -> Result<PoolBlockList, AppError> {
        let response = self.client.clone().list_blocks(req).await?;
        Ok(response.into_inner())
    }

    pub async fn block_category(&self, req: BlockPoolCategoryRequest) -> Result<(), AppError> {
        self.client.clone().block_category(req).await?;
        Ok(())
    }

    pub async fn unblock_category(&self, req: UnblockPoolCategoryRequest) -> Result<(), AppError> {
        self.client.clone().unblock_category(req).await?;
        Ok(())
    }

    pub async fn get_credits(&self, req: GetCreditsRequest) -> Result<CreditsResponse, AppError> {
        let response = self.client.clone().get_credits(req).await?;
        Ok(response.into_inner())
    }
}
