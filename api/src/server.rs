pub mod proto {
    tonic::include_proto!("proto");
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("Proto module loaded");
    println!("");
    Ok(())
}

