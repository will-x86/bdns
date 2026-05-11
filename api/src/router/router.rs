use axum::{Router, routing::post};

use crate::router::{
    AppState,
    routes::{
        auth::{login, sign_up},
        category::{block_category, list_blocked},
    },
};

pub fn app(state: AppState) -> Router {
    let auth_router = Router::new()
        .route("/sign_up", post(sign_up))
        .route("/login", post(login));
    let category_router = Router::new()
        .route("/list-blocked", post(list_blocked))
        .route("/block-category", post(block_category));
    let api_router = Router::new()
        .nest("/auth/", auth_router)
        .nest("/category", category_router);
    Router::new().nest("/api", api_router).with_state(state)
}
