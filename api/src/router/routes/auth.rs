use axum::{Json, extract::State};

use crate::{
    proto::proto::{LoginRequest, LoginResponse, SignUpRequest, SignUpResponse},
    router::{AppError, AppState},
};
pub async fn sign_up(
    State(state): State<AppState>,
    Json(body): Json<SignUpRequest>,
) -> Result<Json<SignUpResponse>, AppError> {
    let resp = state.auth_svc.sign_up(body.timezone).await?;

    Ok(Json(resp))
}

pub async fn login(
    State(state): State<AppState>,
    Json(body): Json<LoginRequest>,
) -> Result<Json<LoginResponse>, AppError> {
    let resp = state.auth_svc.login(body.user_id).await?;
    Ok(Json(resp))
}
