use axum::{Router, routing::post};

use crate::router::{
    AppState,
    routes::{
        auth::{login, sign_up},
        category::{block_category, list_blocked},
        pool::{
            create_pool, delete_pool, get_credits, get_pool, join_pool, leave_pool, list_blocks,
            list_members, list_pools, pool_block_category, pool_unblock_category,
        },
        profile::{create_profile, delete_profile, get_profile, list_profiles, update_profile},
    },
};

pub fn app(state: AppState) -> Router {
    let auth_router = Router::new()
        .route("/sign_up", post(sign_up))
        .route("/login", post(login));
    let category_router = Router::new()
        .route("/list-blocked", post(list_blocked))
        .route("/block-category", post(block_category));
    let pool_router = Router::new()
        .route("/list", post(list_pools))
        .route("/create", post(create_pool))
        .route("/get", post(get_pool))
        .route("/delete", post(delete_pool))
        .route("/join", post(join_pool))
        .route("/leave", post(leave_pool))
        .route("/members", post(list_members))
        .route("/blocks", post(list_blocks))
        .route("/block-category", post(pool_block_category))
        .route("/unblock-category", post(pool_unblock_category))
        .route("/credits", post(get_credits));
    let profile_router = Router::new()
        .route("/list", post(list_profiles))
        .route("/create", post(create_profile))
        .route("/get", post(get_profile))
        .route("/update", post(update_profile))
        .route("/delete", post(delete_profile));
    let api_router = Router::new()
        .nest("/auth", auth_router)
        .nest("/category", category_router)
        .nest("/pool", pool_router)
        .nest("/profile", profile_router);
    Router::new().nest("/api", api_router).with_state(state)
}
