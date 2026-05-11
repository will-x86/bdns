fn main() -> Result<(), Box<dyn std::error::Error>> {
    tonic_prost_build::configure()
        .type_attribute(".", "#[derive(serde::Serialize, serde::Deserialize)]")
        .build_server(false)
        .compile_protos(
            &[
                "proto/auth.proto",
                "proto/user.proto",
                "proto/profile.proto",
                "proto/whitelist.proto",
                "proto/category.proto",
                "proto/timeblock.proto",
                "proto/pool.proto",
            ],
            &["proto"],
        )?;
    Ok(())
}
