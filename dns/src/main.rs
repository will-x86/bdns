#![cfg_attr(not(feature = "openssl"), allow(unused))]

use std::any::Any;
use std::sync::Arc;

use async_trait::async_trait;
use clap::Parser;
use http::header::{CONTENT_LENGTH, CONTENT_TYPE};
use http::{Response, StatusCode};
use pingora::apps::http_app::ServeHttp;
use pingora::listeners::TlsAccept;
use pingora::listeners::tls::TlsSettings;
use pingora::prelude::*;
use pingora::protocols::http::ServerSession;
use pingora::protocols::tls::TlsRef;
use pingora::services::listening::Service;
#[cfg(feature = "openssl")]
use pingora_openssl::ssl::{NameType, SslFiletype};

// Custom structure to hold TLS information
struct MyTlsInfo {
    // SNI (Server Name Indication) from the TLS handshake
    sni: Option<String>,
}

struct MyApp;

#[async_trait]
impl ServeHttp for MyApp {
    async fn response(&self, session: &mut ServerSession) -> http::Response<Vec<u8>> {
        // Extract TLS info from the session's digest extensions
        let my_tls_info = session
            .digest()
            .and_then(|digest| digest.ssl_digest.as_ref())
            .and_then(|ssl_digest| ssl_digest.extension.get::<MyTlsInfo>());
        let sni = my_tls_info
            .and_then(|my_tls_info| my_tls_info.sni.as_deref())
            .unwrap_or("<none>");

        let mut message = String::new();
        message += &format!("Your SNI was: {sni}\n");
        let message = message.into_bytes();

        Response::builder()
            .status(StatusCode::OK)
            .header(CONTENT_TYPE, "text/plain")
            .header(CONTENT_LENGTH, message.len())
            .body(message)
            .unwrap()
    }
}

struct MyTlsCallbacks;

#[async_trait]
impl TlsAccept for MyTlsCallbacks {
    #[cfg(feature = "openssl")]
    async fn handshake_complete_callback(
        &self,
        tls_ref: &TlsRef,
    ) -> Option<Arc<dyn Any + Send + Sync>> {
        // Here you can inspect the TLS connection and return an extension if needed.

        // Extract SNI (Server Name Indication)
        let sni = tls_ref
            .servername(NameType::HOST_NAME)
            .map(ToOwned::to_owned);

        let tls_info = MyTlsInfo { sni };
        Some(Arc::new(tls_info))
    }
}

// This example demonstrates an HTTP server that requires client certificates.
// The server extracts the SNI (Server Name Indication) from the TLS handshake
#[cfg(feature = "openssl")]
fn main() -> Result<(), Box<dyn std::error::Error>> {
    use pingora_openssl::{ssl::SslVerifyMode, x509::X509Name};

    env_logger::init();

    // read command line arguments
    let opt = Opt::parse();

    let mut my_server = Server::new(Some(opt))?;
    my_server.bootstrap();

    let mut my_app = Service::new("my app".to_owned(), MyApp);

    // Paths to server certificate, private key, and client CA certificate
    let manifest_dir = env!("CARGO_MANIFEST_DIR");
    let server_cert_path = format!("{manifest_dir}/keys/server/cert.pem");
    let server_key_path = format!("{manifest_dir}/keys/server/key.pem");
    let client_ca_path = format!("{manifest_dir}/keys/client-ca/cert.pem");

    // Create TLS settings with callbacks
    let callbacks = Box::new(MyTlsCallbacks);
    let mut tls_settings = TlsSettings::with_callbacks(callbacks)?;
    // Set server certificate and private key
    tls_settings.set_certificate_chain_file(&server_cert_path)?;
    tls_settings.set_private_key_file(server_key_path, SslFiletype::PEM)?;
    // Require client certificate
    tls_settings.set_verify(SslVerifyMode::PEER | SslVerifyMode::FAIL_IF_NO_PEER_CERT);
    // Set CA for client certificate verification
    tls_settings.set_ca_file(&client_ca_path)?;
    // Optionally, set the list of acceptable client CAs sent to the client
    tls_settings.set_client_ca_list(X509Name::load_client_ca_file(&client_ca_path)?);

    my_app.add_tls_with_settings("0.0.0.0:6196", None, tls_settings);
    my_server.add_service(my_app);

    my_server.run_forever();
}
