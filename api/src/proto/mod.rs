pub mod proto {
    tonic::include_proto!("proto");
}

pub mod auth;
pub mod user;

pub use auth::AuthSvc;
pub use user::UserSvc;