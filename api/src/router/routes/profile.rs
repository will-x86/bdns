use axum::{Json, extract::State};

use crate::{
    proto::proto::{
        CreateProfileRequest, DeleteProfileRequest, GetProfileRequest, ListProfilesRequest,
        Profile, UpdateProfileRequest,
    },
    router::{AppError, AppState},
};

pub async fn list_profiles(
    State(state): State<AppState>,
    Json(body): Json<ListProfilesRequest>,
) -> Result<Json<Vec<Profile>>, AppError> {
    let resp = state.profile_svc.list_profiles(body.user_id).await?;
    Ok(Json(resp))
}

pub async fn create_profile(
    State(state): State<AppState>,
    Json(body): Json<CreateProfileRequest>,
) -> Result<Json<Profile>, AppError> {
    let resp = state
        .profile_svc
        .create_profile(body.user_id, body.name)
        .await?;
    Ok(Json(resp))
}

pub async fn get_profile(
    State(state): State<AppState>,
    Json(body): Json<GetProfileRequest>,
) -> Result<Json<Profile>, AppError> {
    let resp = state.profile_svc.get_profile(body.profile_id).await?;
    Ok(Json(resp))
}

pub async fn update_profile(
    State(state): State<AppState>,
    Json(body): Json<UpdateProfileRequest>,
) -> Result<Json<Profile>, AppError> {
    let resp = state
        .profile_svc
        .update_profile(body.profile_id, body.name)
        .await?;
    Ok(Json(resp))
}

pub async fn delete_profile(
    State(state): State<AppState>,
    Json(body): Json<DeleteProfileRequest>,
) -> Result<Json<()>, AppError> {
    state.profile_svc.delete_profile(body.profile_id).await?;
    Ok(Json(()))
}
