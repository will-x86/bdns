use axum::{Json, extract::State};

use crate::{
    proto::proto::{
        BlockCategoryRequest, CategoryBlock, ListBlockedRequest, UnblockCategoryRequest,
    },
    router::{AppError, AppState},
};
pub async fn list_blocked(
    State(state): State<AppState>,
    Json(body): Json<ListBlockedRequest>,
) -> Result<Json<Vec<String>>, AppError> {
    let resp = state.category_svc.list_blocked(body.profile_id).await?;

    Ok(Json(resp))
}

pub async fn block_category(
    State(state): State<AppState>,
    Json(body): Json<BlockCategoryRequest>,
) -> Result<Json<CategoryBlock>, AppError> {
    let resp = state
        .category_svc
        .block_category(body.profile_id, body.category)
        .await?;
    Ok(Json(resp))
}

pub async fn unblock_category(
    State(state): State<AppState>,
    Json(body): Json<UnblockCategoryRequest>,
) -> Result<Json<()>, AppError> {
    let resp = state
        .category_svc
        .unblock_category(body.profile_id, body.category)
        .await?;
    Ok(Json(resp))
}
