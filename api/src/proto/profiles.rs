use crate::{
    proto::proto::{
        CreateProfileRequest, DeleteProfileRequest, GetProfileRequest, ListProfilesRequest,
        ListProfilesResponse, Profile, UpdateProfileRequest,
    },
    router::AppError,
};
use tonic::transport::Channel;

use super::proto::profile_service_client::ProfileServiceClient;

#[derive(Clone)]
pub struct ProfileSvc {
    client: ProfileServiceClient<Channel>,
}

impl ProfileSvc {
    pub async fn new(addr: String) -> Result<Self, AppError> {
        let client = ProfileServiceClient::connect(addr).await?;
        Ok(Self { client })
    }

    pub async fn list_profiles(&self, req: ListProfilesRequest) -> Result<ListProfilesResponse, AppError> {
        let response = self.client.clone().list_profiles(req).await?;
        Ok(response.into_inner())
    }

    pub async fn create_profile(&self, req: CreateProfileRequest) -> Result<Profile, AppError> {
        let response = self.client.clone().create_profile(req).await?;
        Ok(response.into_inner())
    }

    pub async fn get_profile(&self, req: GetProfileRequest) -> Result<Profile, AppError> {
        let response = self.client.clone().get_profile(req).await?;
        Ok(response.into_inner())
    }

    pub async fn update_profile(&self, req: UpdateProfileRequest) -> Result<Profile, AppError> {
        let response = self.client.clone().update_profile(req).await?;
        Ok(response.into_inner())
    }

    pub async fn delete_profile(&self, req: DeleteProfileRequest) -> Result<(), AppError> {
        self.client.clone().delete_profile(req).await?;
        Ok(())
    }
}
