use sqlx::{QueryBuilder, Any};
use crate::evaluator::Value;

#[derive(Clone, Debug)]
pub struct ZenoQueryBuilder {
    table: String,
    select: Vec<String>,
    wheres: Vec<(String, String, Value)>, // col, op, val
    order_by: Vec<(String, String)>, // col, direction
    limit: Option<i64>,
    offset: Option<i64>,
}

impl ZenoQueryBuilder {
    pub fn new(table: String) -> Self {
        Self {
            table,
            select: vec!["*".to_string()],
            wheres: Vec::new(),
            order_by: Vec::new(),
            limit: None,
            offset: None,
        }
    }

    pub fn select(&mut self, fields: Vec<String>) {
        self.select = fields;
    }

    pub fn where_clause(&mut self, col: String, op: String, val: Value) {
        self.wheres.push((col, op, val));
    }

    pub fn order_by(&mut self, col: String, dir: String) {
        self.order_by.push((col, dir));
    }

    pub fn limit(&mut self, limit: i64) {
        self.limit = Some(limit);
    }

    pub fn offset(&mut self, offset: i64) {
        self.offset = Some(offset);
    }

    pub fn to_sql_and_params(&self) -> (String, Vec<Value>) {
        let mut sql = String::from("SELECT ");
        sql.push_str(&self.select.join(", "));
        sql.push_str(" FROM ");
        sql.push_str(&self.table);

        let mut params = Vec::new();

        if !self.wheres.is_empty() {
            sql.push_str(" WHERE ");
            for (i, (col, op, val)) in self.wheres.iter().enumerate() {
                if i > 0 { sql.push_str(" AND "); }
                sql.push_str(col);
                sql.push_str(" ");
                sql.push_str(op);
                sql.push_str(" ?");
                params.push(val.clone());
            }
        }

        if !self.order_by.is_empty() {
            sql.push_str(" ORDER BY ");
            let orders: Vec<String> = self.order_by.iter()
                .map(|(c, d)| format!("{} {}", c, d))
                .collect();
            sql.push_str(&orders.join(", "));
        }

        if let Some(l) = self.limit {
            sql.push_str(" LIMIT ?");
            params.push(Value::Integer(l));
        }

        if let Some(o) = self.offset {
            sql.push_str(" OFFSET ?");
            params.push(Value::Integer(o));
        }

        (sql, params)
    }
}
