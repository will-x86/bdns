use crate::proto::proto::{LoginResponse, SignUpResponse};

use super::proto::auth_client::AuthClient;
use tonic::transport::Channel;

pub struct AuthSvc {
    client: AuthClient<Channel>,
}

impl AuthSvc {
    pub async fn new(addr: String) -> Result<Self, Box<dyn std::error::Error>> {
        let client = AuthClient::connect(addr).await?;
        Ok(Self { client })
    }

    pub async fn sign_up(
        &mut self,
        timezone: String,
    ) -> Result<SignUpResponse, Box<dyn std::error::Error>> {
        use super::proto::SignUpRequest;

        let request = SignUpRequest { timezone };

        let response = self.client.sign_up(request).await?;
        let inner = response.into_inner();

        Ok(inner)
    }

    pub async fn login(
        &mut self,
        user_id: String,
    ) -> Result<LoginResponse, Box<dyn std::error::Error>> {
        use super::proto::LoginRequest;

        let request = LoginRequest { user_id };

        let response = self.client.login(request).await?;
        let inner = response.into_inner();

        Ok(inner)
    }
}

