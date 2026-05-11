use axum::{Router, routing::post};

use crate::router::{
    AppState,
    routes::{
        auth::{login, sign_up},
        category::{block_category, list_blocked, unblock_category},
        pool::{
            create_pool, delete_pool, get_credits, get_pool, join_pool, leave_pool, list_blocks,
            list_members, list_pools, pool_block_category, pool_unblock_category,
        },
        profile::{create_profile, delete_profile, get_profile, list_profiles, update_profile},
        timeblock::{create_timeblock, delete_timeblock, list_timeblocks},
        user::{get_user, update_user},
        whitelist::{
            add_permanent, add_temporary, list_permanent, list_temporary, remove_permanent,
            remove_temporary,
        },
    },
};

pub fn app(state: AppState) -> Router {
    let auth_router = Router::new()
        .route("/sign_up", post(sign_up))
        .route("/login", post(login));
    let category_router = Router::new()
        .route("/list-blocked", post(list_blocked))
        .route("/block-category", post(block_category))
        .route("/unblock-category", post(unblock_category));
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
    let timeblock_router = Router::new()
        .route("/list", post(list_timeblocks))
        .route("/create", post(create_timeblock))
        .route("/delete", post(delete_timeblock));
    let user_router = Router::new()
        .route("/get", post(get_user))
        .route("/update", post(update_user));
    let whitelist_router = Router::new()
        .route("/list-permanent", post(list_permanent))
        .route("/add-permanent", post(add_permanent))
        .route("/remove-permanent", post(remove_permanent))
        .route("/list-temporary", post(list_temporary))
        .route("/add-temporary", post(add_temporary))
        .route("/remove-temporary", post(remove_temporary));
    let api_router = Router::new()
        .nest("/auth", auth_router)
        .nest("/category", category_router)
        .nest("/pool", pool_router)
        .nest("/profile", profile_router)
        .nest("/timeblock", timeblock_router)
        .nest("/user", user_router)
        .nest("/whitelist", whitelist_router);
    Router::new().nest("/api", api_router).with_state(state)
}
