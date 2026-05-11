use crate::{
    proto::proto::{
        AddPermanentRequest, AddTemporaryRequest, ListPermanentRequest, ListTemporaryRequest,
        RemovePermanentRequest, RemoveTemporaryRequest, WhitelistDomain, WhitelistDomainTemp,
    },
    router::{AppError, AppState},
};
use axum::{Json, extract::State};

pub async fn list_permanent(
    State(state): State<AppState>,
    Json(body): Json<ListPermanentRequest>,
) -> Result<Json<Vec<String>>, AppError> {
    let resp = state.whitelist_svc.list_permanent(body.profile_id).await?;
    Ok(Json(resp))
}

pub async fn add_permanent(
    State(state): State<AppState>,
    Json(body): Json<AddPermanentRequest>,
) -> Result<Json<WhitelistDomain>, AppError> {
    let resp = state
        .whitelist_svc
        .add_permanent(body.profile_id, body.domain)
        .await?;
    Ok(Json(resp))
}

pub async fn remove_permanent(
    State(state): State<AppState>,
    Json(body): Json<RemovePermanentRequest>,
) -> Result<(), AppError> {
    state
        .whitelist_svc
        .remove_permanent(body.profile_id, body.domain)
        .await?;
    Ok(())
}

pub async fn list_temporary(
    State(state): State<AppState>,
    Json(body): Json<ListTemporaryRequest>,
) -> Result<Json<Vec<WhitelistDomainTemp>>, AppError> {
    let resp = state.whitelist_svc.list_temporary(body.profile_id).await?;
    Ok(Json(resp))
}

pub async fn add_temporary(
    State(state): State<AppState>,
    Json(body): Json<AddTemporaryRequest>,
) -> Result<Json<WhitelistDomainTemp>, AppError> {
    let resp = state
        .whitelist_svc
        .add_temporary(body.profile_id, body.domain, body.expires_at)
        .await?;
    Ok(Json(resp))
}

pub async fn remove_temporary(
    State(state): State<AppState>,
    Json(body): Json<RemoveTemporaryRequest>,
) -> Result<(), AppError> {
    state
        .whitelist_svc
        .remove_temporary(body.profile_id, body.domain)
        .await?;
    Ok(())
}
