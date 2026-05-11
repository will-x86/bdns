use axum::{Json, extract::State};

use crate::{
    proto::proto::{
        BlockPoolCategoryRequest, CreatePoolRequest, DeletePoolRequest, FriendPool, GetPoolRequest,
        JoinPoolRequest, LeavePoolRequest, ListMembersRequest, ListPoolBlocksRequest,
        ListPoolsRequest, PoolBlock, PoolMember, UnblockPoolCategoryRequest,
    },
    router::{AppError, AppState},
};

pub async fn list_pools(
    State(state): State<AppState>,
    Json(body): Json<ListPoolsRequest>,
) -> Result<Json<Vec<FriendPool>>, AppError> {
    let resp = state.pool_svc.list_pools(body.user_id).await?;
    Ok(Json(resp))
}

pub async fn create_pool(
    State(state): State<AppState>,
    Json(body): Json<CreatePoolRequest>,
) -> Result<Json<FriendPool>, AppError> {
    let resp = state
        .pool_svc
        .create_pool(body.user_id, body.name, body.pool_mode, body.total_limit)
        .await?;
    Ok(Json(resp))
}

pub async fn get_pool(
    State(state): State<AppState>,
    Json(body): Json<GetPoolRequest>,
) -> Result<Json<FriendPool>, AppError> {
    let resp = state.pool_svc.get_pool(body.pool_id).await?;
    Ok(Json(resp))
}

pub async fn delete_pool(
    State(state): State<AppState>,
    Json(body): Json<DeletePoolRequest>,
) -> Result<Json<()>, AppError> {
    state
        .pool_svc
        .delete_pool(body.pool_id, body.user_id)
        .await?;
    Ok(Json(()))
}

pub async fn join_pool(
    State(state): State<AppState>,
    Json(body): Json<JoinPoolRequest>,
) -> Result<Json<()>, AppError> {
    state
        .pool_svc
        .join_pool(body.pool_id, body.profile_id)
        .await?;
    Ok(Json(()))
}

pub async fn leave_pool(
    State(state): State<AppState>,
    Json(body): Json<LeavePoolRequest>,
) -> Result<Json<()>, AppError> {
    state
        .pool_svc
        .leave_pool(body.pool_id, body.profile_id)
        .await?;
    Ok(Json(()))
}

pub async fn list_members(
    State(state): State<AppState>,
    Json(body): Json<ListMembersRequest>,
) -> Result<Json<Vec<PoolMember>>, AppError> {
    let resp = state.pool_svc.list_members(body.pool_id).await?;
    Ok(Json(resp))
}

pub async fn list_blocks(
    State(state): State<AppState>,
    Json(body): Json<ListPoolBlocksRequest>,
) -> Result<Json<Vec<PoolBlock>>, AppError> {
    let resp = state.pool_svc.list_blocks(body.pool_id).await?;
    Ok(Json(resp))
}

pub async fn pool_block_category(
    State(state): State<AppState>,
    Json(body): Json<BlockPoolCategoryRequest>,
) -> Result<Json<()>, AppError> {
    state
        .pool_svc
        .block_category(body.pool_id, body.category)
        .await?;
    Ok(Json(()))
}

pub async fn pool_unblock_category(
    State(state): State<AppState>,
    Json(body): Json<UnblockPoolCategoryRequest>,
) -> Result<Json<()>, AppError> {
    state
        .pool_svc
        .unblock_category(body.pool_id, body.category)
        .await?;
    Ok(Json(()))
}

pub async fn get_credits(
    State(state): State<AppState>,
    Json(body): Json<crate::proto::proto::GetCreditsRequest>,
) -> Result<Json<crate::proto::proto::CreditsResponse>, AppError> {
    let resp = state
        .pool_svc
        .get_credits(body.pool_id, body.profile_id)
        .await?;
    Ok(Json(resp))
}
