use axum::Json;
use axum::extract::State;

use crate::proto::proto::{
    CreateTimeBlockRequest, DeleteTimeBlockRequest, ListTimeBlocksRequest, TimeBlock,
};
use crate::router::{AppError, AppState};

pub async fn list_timeblocks(
    State(state): State<AppState>,
    Json(body): Json<ListTimeBlocksRequest>,
) -> Result<Json<Vec<TimeBlock>>, AppError> {
    let resp = state.timeblock_svc.list(body.profile_id).await?;
    Ok(Json(resp))
}

pub async fn create_timeblock(
    State(state): State<AppState>,
    Json(body): Json<CreateTimeBlockRequest>,
) -> Result<Json<TimeBlock>, AppError> {
    let resp = state
        .timeblock_svc
        .create(
            body.profile_id,
            body.category,
            body.start_time,
            body.end_time,
            body.day,
        )
        .await?;
    Ok(Json(resp))
}

pub async fn delete_timeblock(
    State(state): State<AppState>,
    Json(body): Json<DeleteTimeBlockRequest>,
) -> Result<Json<()>, AppError> {
    state.timeblock_svc.delete(body.block_id).await?;
    Ok(Json(()))
}
