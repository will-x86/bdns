use axum::{
    http::StatusCode,
    response::{IntoResponse, Response},
};

use crate::proto::{
    AuthSvc, CategorySvc, PoolSvc, ProfileSvc, TimeBlockSvc, UserSvc, WhitelistSvc,
};

pub mod router;
pub mod routes;

#[derive(Clone)]
pub struct AppState {
    pub auth_svc: AuthSvc,
    pub category_svc: CategorySvc,
    pub pool_svc: PoolSvc,
    pub profile_svc: ProfileSvc,
    pub timeblock_svc: TimeBlockSvc,
    pub user_svc: UserSvc,
    pub whitelist_svc: WhitelistSvc,
}
#[derive(Debug)]
pub struct AppError(anyhow::Error);

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            format!("Something went wrong: {}", self.0),
        )
            .into_response()
    }
}

impl<E> From<E> for AppError
where
    E: Into<anyhow::Error>,
{
    fn from(err: E) -> Self {
        Self(err.into())
    }
}
