pub mod proto {
    tonic::include_proto!("proto");
}

pub mod auth;
pub mod user;
pub mod profiles;

pub use auth::AuthSvc;
pub use user::UserSvc;
pub use profiles::ProfileSvc;