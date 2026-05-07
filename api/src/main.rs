mod proto;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let mut auth_svc = proto::AuthSvc::new("http://[::1]:50051".to_string()).await?;

    let user = auth_svc.sign_up("Europe/London".to_string()).await?;
    println!(
        "Signed up: user_id={}, created_at={}",
        user.user_id, user.created_at
    );

    let success = auth_svc.login(user.user_id.clone()).await?;
    println!("Login success: {}", success.user_id == user.user_id);

    let mut user_svc = proto::UserSvc::new("http://[::1]:50051".to_string()).await?;
    let user = user_svc.get_user(user.user_id.clone()).await?;
    println!(
        "Got user: id={}, timezone={}, created_at={}",
        user.id, user.timezone, user.created_at
    );

    let user = user_svc
        .update_user(user.id, "America/New_York".to_string())
        .await?;
    println!(
        "Updated user: id={}, timezone={}, created_at={}",
        user.id, user.timezone, user.created_at
    );

    let mut profile_svc = proto::ProfileSvc::new("http://[::1]:50051".to_string()).await?;

    let profile = profile_svc
        .create_profile(user.id.clone(), "laptop".to_string())
        .await?;
    println!(
        "Created profile: id={}, user_id={}, name={}",
        profile.id, profile.user_id, profile.name
    );

    let profiles = profile_svc.list_profiles(user.id.clone()).await?;
    println!("Listed {} profiles", profiles.len());

    let profile = profile_svc
        .update_profile(profile.id.clone(), "desktop".to_string())
        .await?;
    println!(
        "Updated profile: id={}, name={}",
        profile.id, profile.name
    );

    let profile = profile_svc.get_profile(profile.id.clone()).await?;
    println!(
        "Got profile: id={}, name={}",
        profile.id, profile.name
    );

    profile_svc.delete_profile(profile.id.clone()).await?;
    println!("Deleted profile: id={}", profile.id);

    Ok(())
}

