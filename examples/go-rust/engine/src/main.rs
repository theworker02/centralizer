// Minimal Centralizer Protocol 1.x speaker. No Centralizer internals.
use std::io::{self, BufRead, Write};

fn encode_float(v: f64) -> String {
    format!(r#"{{"k":"float","f":{}}}"#, v)
}

fn decode_number(args: &str) -> f64 {
    // Prefer the first "f" or "i" field in the payload args object.
    for key in ["\"f\":", "\"i\":"] {
        if let Some(idx) = args.find(key) {
            let rest = &args[idx + key.len()..];
            let num: String = rest
                .chars()
                .take_while(|c| c.is_ascii_digit() || *c == '.' || *c == '-')
                .collect();
            if let Ok(v) = num.parse::<f64>() {
                return v;
            }
        }
    }
    0.0
}

fn send(id: &str, typ: &str, payload: &str) {
    let mut out = io::stdout();
    let line = format!(
        r#"{{"v":1,"id":"{}","type":"{}","payload":{}}}"#,
        id, typ, payload
    );
    writeln!(out, "{}", line).ok();
    out.flush().ok();
}

fn main() {
    let stdin = io::stdin();
    for line in stdin.lock().lines() {
        let line = match line {
            Ok(l) => l,
            Err(_) => break,
        };
        if line.trim().is_empty() {
            continue;
        }
        let id = extract(&line, "\"id\":\"", '"').unwrap_or_else(|| "0".into());
        let typ = extract(&line, "\"type\":\"", '"').unwrap_or_default();
        match typ.as_str() {
            "HELLO" => send(
                &id,
                "HELLO",
                r#"{"protocol":"1.0","name":"centralizer-rust","features":["call"],"runtime":"rustc"}"#,
            ),
            "DESCRIBE" => send(
                &id,
                "DESCRIBE_OK",
                r#"{"schema":"service: engine\ninferred: true\nfunctions:\n  multiply:\n    args:\n      value: float\n    returns:\n      type: float\n"}"#,
            ),
            "CALL" => {
                let value = decode_number(&line);
                let result = value * 2.0;
                send(
                    &id,
                    "RESULT",
                    &format!(r#"{{"value":{}}}"#, encode_float(result)),
                );
            }
            "HEARTBEAT" | "CANCEL" => send(&id, "OK", "{}"),
            "SHUTDOWN" => {
                send(&id, "OK", "{}");
                break;
            }
            _ => send(
                &id,
                "ERROR",
                r#"{"code":"protocol","message":"unsupported"}"#,
            ),
        }
    }
}

fn extract(s: &str, start: &str, end: char) -> Option<String> {
    let i = s.find(start)?;
    let rest = &s[i + start.len()..];
    let j = rest.find(end)?;
    Some(rest[..j].to_string())
}
