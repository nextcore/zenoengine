pub mod math;
pub mod date;

use zenocore::Engine;

pub fn register_std_slots(engine: &mut Engine) {
    math::register(engine);
    date::register(engine);
}

// ==========================================
// SHAPED CHRONO HELPER FUNCTIONS
// ==========================================

pub(crate) fn translate_layout_to_chrono(layout: &str) -> String {
    let layout_lower = layout.to_lowercase();
    match layout_lower.as_str() {
        "human" => return "%d %b %Y %H:%M".to_string(),
        "date" => return "%Y-%m-%d".to_string(),
        "time" => return "%H:%M:%S".to_string(),
        "rfc3339" => return "%Y-%m-%dT%H:%M:%S%:z".to_string(),
        "full" => return "%A, %d %B %Y %H:%M:%S".to_string(),
        _ => {}
    }

    let mut res = layout.to_string();
    res = res.replace("2006-01-02 15:04:05", "%Y-%m-%d %H:%M:%S");
    res = res.replace("2006-01-02", "%Y-%m-%d");
    res = res.replace("15:04:05", "%H:%M:%S");
    res = res.replace("2006", "%Y");
    res = res.replace("06", "%y");
    res = res.replace("01", "%m");
    res = res.replace("02", "%d");
    res = res.replace("15", "%H");
    res = res.replace("04", "%M");
    res = res.replace("05", "%S");

    res = res.replace("yyyy", "%Y");
    res = res.replace("yy", "%y");
    res = res.replace("MMMM", "%B");
    res = res.replace("MMM", "%b");
    res = res.replace("MM", "%m");
    res = res.replace("dddd", "%A");
    res = res.replace("ddd", "%a");
    res = res.replace("dd", "%d");
    res = res.replace("HH", "%H");
    res = res.replace("hh", "%I");
    res = res.replace("mm", "%M");
    res = res.replace("ss", "%S");
    res = res.replace("tt", "%p");

    res
}

pub(crate) fn parse_flex_date(s: &str) -> Option<chrono::DateTime<chrono::Local>> {
    use chrono::TimeZone;
    if let Ok(dt) = chrono::DateTime::parse_from_rfc3339(s) {
        return Some(dt.with_timezone(&chrono::Local));
    }
    let formats = [
        "%Y-%m-%d %H:%M:%S",
        "%Y-%m-%d",
        "%d-%m-%Y",
        "%d/%m/%Y",
        "%d %b %Y",
    ];
    for fmt in &formats {
        if fmt.contains("%H") {
            if let Ok(ndt) = chrono::NaiveDateTime::parse_from_str(s, fmt) {
                if let Some(dt) = chrono::Local.from_local_datetime(&ndt).earliest() {
                    return Some(dt);
                }
            }
        } else {
            if let Ok(nd) = chrono::NaiveDate::parse_from_str(s, fmt) {
                if let Some(ndt) = nd.and_hms_opt(0, 0, 0) {
                    if let Some(dt) = chrono::Local.from_local_datetime(&ndt).earliest() {
                        return Some(dt);
                    }
                }
            }
        }
    }
    None
}

pub(crate) fn shift_datetime(dt: &chrono::DateTime<chrono::Local>, dur_str: &str) -> Option<chrono::DateTime<chrono::Local>> {
    let s = dur_str.trim();
    if s.is_empty() {
        return Some(*dt);
    }

    let (is_negative, rest) = if s.starts_with('-') {
        (true, &s[1..])
    } else if s.starts_with('+') {
        (false, &s[1..])
    } else {
        (false, s)
    };

    let mut num_str = String::new();
    let mut unit = String::new();
    for c in rest.chars() {
        if c.is_ascii_digit() {
            num_str.push(c);
        } else {
            unit.push(c);
        }
    }

    let val = num_str.parse::<i64>().ok()?;
    let val_signed = if is_negative { -val } else { val };

    match unit.trim().to_lowercase().as_str() {
        "s" => Some(*dt + chrono::Duration::seconds(val_signed)),
        "m" | "min" => Some(*dt + chrono::Duration::minutes(val_signed)),
        "h" | "hr" | "hour" | "hours" => Some(*dt + chrono::Duration::hours(val_signed)),
        "d" | "day" | "days" => Some(*dt + chrono::Duration::days(val_signed)),
        _ => None,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use zenocore::{Engine, Scope, Value, parse_string, register_logic_slots, Context};

    #[test]
    fn test_math_calc_slot() {
        let mut engine = Engine::new();
        register_logic_slots(&mut engine);
        register_std_slots(&mut engine);

        let code = r#"
            var: $total {
              val: 25
            }
            math.calc: "ceil($total / 10)" {
              as: $pages
            }
            math.calc: "pow(2, 3)" {
              as: $powered
            }
            math.calc: "max(1, 5, 3)" {
              as: $maximum
            }
        "#;
        let root = parse_string(code, "test.zl").unwrap();
        let mut ctx = Context::new();
        let scope = Scope::new(None);
        engine.execute(&mut ctx, &root, &scope).unwrap();

        assert_eq!(scope.get("pages").unwrap(), Value::Float(3.0));
        assert_eq!(scope.get("powered").unwrap(), Value::Float(8.0));
        assert_eq!(scope.get("maximum").unwrap(), Value::Float(5.0));
    }

    #[test]
    fn test_money_calc_slot() {
        let mut engine = Engine::new();
        register_logic_slots(&mut engine);
        register_std_slots(&mut engine);

        let code = r#"
            var: $harga {
              val: "100.50"
            }
            var: $qty {
              val: 3
            }
            var: $diskon {
              val: "10.25"
            }
            money.calc: "($harga * $qty) - $diskon" {
              as: $total
            }
        "#;
        let root = parse_string(code, "test.zl").unwrap();
        let mut ctx = Context::new();
        let scope = Scope::new(None);
        engine.execute(&mut ctx, &root, &scope).unwrap();

        assert_eq!(scope.get("total").unwrap(), Value::String("291.25".to_string()));
    }

    #[test]
    fn test_date_slots() {
        let mut engine = Engine::new();
        register_logic_slots(&mut engine);
        register_std_slots(&mut engine);

        let code = r#"
            date.now: {
              layout: "yyyy-MM-dd"
              as: $today
            }
            date.parse: "2023-12-25 15:30:00" {
              layout: "yyyy-MM-dd HH:mm:ss"
              as: $parsed
            }
            date.format: $parsed {
              layout: "date"
              as: $formatted
            }
            date.add: $parsed {
              duration: "2h"
              as: $shifted
            }
        "#;
        let root = parse_string(code, "test.zl").unwrap();
        let mut ctx = Context::new();
        let scope = Scope::new(None);
        engine.execute(&mut ctx, &root, &scope).unwrap();

        let today_str = scope.get("today").unwrap().to_string_coerce();
        assert_eq!(today_str.len(), 10);
        assert!(today_str.contains('-'));

        assert_eq!(scope.get("formatted").unwrap(), Value::String("2023-12-25".to_string()));

        let shifted_str = scope.get("shifted").unwrap().to_string_coerce();
        assert!(shifted_str.contains("17:30:00"));
    }
}
