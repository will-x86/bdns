use axum::{Json, extract::State};

use crate::{
    proto::proto::{GetUserRequest, UpdateUserRequest, User},
    router::{AppError, AppState},
};
pub async fn get_user(
    State(state): State<AppState>,
    Json(body): Json<GetUserRequest>,
) -> Result<Json<User>, AppError> {
    let resp = state.user_svc.get_user(body.user_id).await?;

    Ok(Json(resp))
}

pub async fn update_user(
    State(state): State<AppState>,
    Json(body): Json<UpdateUserRequest>,
) -> Result<Json<User>, AppError> {
    let resp = state
        .user_svc
        .update_user(body.user_id, body.timezone)
        .await?;
    Ok(Json(resp))
}
