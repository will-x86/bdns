pub mod proto {
    tonic::include_proto!("proto");
}

pub mod auth;
pub mod category;
pub mod pool;
pub mod profiles;
pub mod timeblock;
pub mod user;
pub mod whitelist;

pub use auth::AuthSvc;
pub use category::CategorySvc;
pub use pool::PoolSvc;
pub use profiles::ProfileSvc;
pub use timeblock::TimeBlockSvc;
pub use user::UserSvc;
pub use whitelist::WhitelistSvc;

