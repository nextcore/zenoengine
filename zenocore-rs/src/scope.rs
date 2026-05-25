use std::collections::HashMap;
use std::sync::{Arc, RwLock};

#[derive(Debug, Clone, PartialEq)]
pub enum Value {
    Nil,
    String(String),
    Int(i64),
    Float(f64),
    Bool(bool),
    List(Vec<Value>),
    Map(HashMap<String, Value>),
}

// Convert from primitives easily
impl From<String> for Value {
    fn from(s: String) -> Self {
        Value::String(s)
    }
}

impl From<&str> for Value {
    fn from(s: &str) -> Self {
        Value::String(s.to_string())
    }
}

impl From<i64> for Value {
    fn from(i: i64) -> Self {
        Value::Int(i)
    }
}

impl From<f64> for Value {
    fn from(f: f64) -> Self {
        Value::Float(f)
    }
}

impl From<bool> for Value {
    fn from(b: bool) -> Self {
        Value::Bool(b)
    }
}

impl Value {
    pub fn to_string_coerce(&self) -> String {
        match self {
            Value::Nil => "".to_string(),
            Value::String(s) => s.clone(),
            Value::Int(i) => i.to_string(),
            Value::Float(f) => f.to_string(),
            Value::Bool(b) => b.to_string(),
            Value::List(l) => format!("{:?}", l),
            Value::Map(m) => format!("{:?}", m),
        }
    }

    pub fn to_bool(&self) -> bool {
        match self {
            Value::Nil => false,
            Value::Bool(b) => *b,
            Value::Int(i) => *i != 0,
            Value::Float(f) => *f != 0.0,
            Value::String(s) => {
                let s_lower = s.to_lowercase();
                s_lower == "true" || s_lower == "1" || s_lower == "yes" || s_lower == "on"
            }
            Value::List(l) => !l.is_empty(),
            Value::Map(m) => !m.is_empty(),
        }
    }

    pub fn to_int(&self) -> i64 {
        match self {
            Value::Nil => 0,
            Value::Bool(b) => if *b { 1 } else { 0 },
            Value::Int(i) => *i,
            Value::Float(f) => *f as i64,
            Value::String(s) => s.parse::<i64>().unwrap_or(0),
            Value::List(_) | Value::Map(_) => 0,
        }
    }

    pub fn to_float(&self) -> f64 {
        match self {
            Value::Nil => 0.0,
            Value::Bool(b) => if *b { 1.0 } else { 0.0 },
            Value::Int(i) => *i as f64,
            Value::Float(f) => *f,
            Value::String(s) => s.parse::<f64>().unwrap_or(0.0),
            Value::List(_) | Value::Map(_) => 0.0,
        }
    }

    pub fn to_list(&self) -> Vec<Value> {
        match self {
            Value::List(l) => l.clone(),
            _ => Vec::new(),
        }
    }

    pub fn to_map(&self) -> HashMap<String, Value> {
        match self {
            Value::Map(m) => m.clone(),
            _ => HashMap::new(),
        }
    }
}

pub struct Scope {
    vars: RwLock<HashMap<String, Value>>,
    parent: RwLock<Option<Arc<Scope>>>,
}

impl Scope {
    pub fn new(parent: Option<Arc<Scope>>) -> Arc<Self> {
        Arc::new(Self {
            vars: RwLock::new(HashMap::new()),
            parent: RwLock::new(parent),
        })
    }

    pub fn set_parent(&self, parent: Option<Arc<Scope>>) {
        let mut p = self.parent.write().unwrap();
        *p = parent;
    }

    pub fn set(&self, key: &str, val: Value) {
        let mut vars = self.vars.write().unwrap();
        vars.insert(key.to_string(), val);
    }

    pub fn delete(&self, key: &str) {
        let mut vars = self.vars.write().unwrap();
        vars.remove(key);
    }

    pub fn get(&self, key: &str) -> Option<Value> {
        // 1. Check direct key in current scope
        {
            let vars = self.vars.read().unwrap();
            if let Some(val) = vars.get(key) {
                return Some(val.clone());
            }
        }

        // 2. Get parent reference to avoid locking current scope while traversing
        let parent = {
            let p = self.parent.read().unwrap();
            p.clone()
        };

        // 3. Check parent scope
        if let Some(ref p) = parent {
            if let Some(val) = p.get(key) {
                return Some(val);
            }
        }

        // 4. Check nested key (Dot Notation Deep Navigation: e.g. "user.name")
        if key.contains('.') {
            let parts: Vec<&str> = key.split('.').collect();
            if parts.is_empty() {
                return None;
            }

            // Find the root part in current scope or parent scope
            let mut current_val = {
                let vars = self.vars.read().unwrap();
                vars.get(parts[0]).cloned()
            };

            if current_val.is_none() {
                if let Some(ref p) = parent {
                    current_val = p.get(parts[0]);
                }
            }

            let mut val = match current_val {
                Some(v) => v,
                None => return None,
            };

            // Traverse the dot paths
            for &part in parts.iter().skip(1) {
                match val {
                    Value::Map(ref map) => {
                        if let Some(next) = map.get(part) {
                            val = next.clone();
                        } else {
                            return None;
                        }
                    }
                    Value::List(ref list) => {
                        if let Ok(idx) = part.parse::<usize>() {
                            if idx < list.len() {
                                val = list[idx].clone();
                            } else {
                                return None;
                            }
                        } else {
                            return None;
                        }
                    }
                    _ => return None,
                }
            }
            return Some(val);
        }

        None
    }

    pub fn get_default(&self, key: &str, default_value: Value) -> Value {
        self.get(key).unwrap_or(default_value)
    }

    pub fn to_map(&self) -> HashMap<String, Value> {
        let vars = self.vars.read().unwrap();
        vars.clone()
    }

    pub fn reset(&self) {
        let mut vars = self.vars.write().unwrap();
        vars.clear();
    }

    pub fn clone_scope(&self) -> Arc<Self> {
        let vars_clone = {
            let vars = self.vars.read().unwrap();
            vars.clone()
        };
        Arc::new(Self {
            vars: RwLock::new(vars_clone),
            parent: RwLock::new(None),
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_scope_basic() {
        let scope = Scope::new(None);
        scope.set("a", Value::from("hello"));
        assert_eq!(scope.get("a"), Some(Value::from("hello")));

        scope.delete("a");
        assert_eq!(scope.get("a"), None);
    }

    #[test]
    fn test_scope_parent() {
        let parent = Scope::new(None);
        parent.set("x", Value::from(42));

        let child = Scope::new(Some(parent));
        assert_eq!(child.get("x"), Some(Value::from(42)));

        child.set("x", Value::from(100));
        assert_eq!(child.get("x"), Some(Value::from(100)));
    }

    #[test]
    fn test_scope_nested() {
        let scope = Scope::new(None);
        
        let mut user_map = HashMap::new();
        user_map.insert("name".to_string(), Value::from("Alice"));
        user_map.insert("age".to_string(), Value::from(30));
        
        scope.set("user", Value::Map(user_map));
        
        assert_eq!(scope.get("user.name"), Some(Value::from("Alice")));
        assert_eq!(scope.get("user.age"), Some(Value::from(30)));
        assert_eq!(scope.get("user.nonexistent"), None);
    }
}
