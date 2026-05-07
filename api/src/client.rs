pub mod proto {
    tonic::include_proto!("proto");
}

use proto::user_service_client::UserServiceClient;
use proto::{GetUserRequest, UpdateUserRequest};

pub async fn run_client() -> Result<(), Box<dyn std::error::Error>> {
    let mut client = UserServiceClient::connect("http://[::1]:50051").await?;

    // Get a user
    let response = client
        .get_user(GetUserRequest {
            user_id: "some-id-123".to_string(),
        })
        .await?;

    let user = response.into_inner();
    println!("Got user -> id: {}, timezone: {}", user.id, user.timezone);

    // Update a user
    let response = client
        .update_user(UpdateUserRequest {
            user_id: "some-id-123".to_string(),
            timezone: "Europe/London".to_string(),
        })
        .await?;

    let user = response.into_inner();
    println!(
        "Updated user -> id: {}, timezone: {}",
        user.id, user.timezone
    );

    Ok(())
}

