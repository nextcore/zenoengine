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

    // Constructs a sqlx::QueryBuilder which handles driver-specific syntax (e.g. $1 vs ?)
    pub fn build_query<'a>(&'a self) -> QueryBuilder<'a, Any> {
        let mut qb = QueryBuilder::new("SELECT ");

        // Fields
        qb.push(self.select.join(", "));

        // From
        qb.push(" FROM ");
        qb.push(&self.table);

        // Where
        if !self.wheres.is_empty() {
            qb.push(" WHERE ");
            for (i, (col, op, val)) in self.wheres.iter().enumerate() {
                if i > 0 {
                    qb.push(" AND ");
                }
                qb.push(col);
                qb.push(" ");
                qb.push(op);
                qb.push(" ");
                // Bind value based on type
                match val {
                    Value::Integer(v) => { qb.push_bind(*v); },
                    Value::String(v) => { qb.push_bind(v.clone()); },
                    Value::Boolean(v) => { qb.push_bind(*v); },
                    Value::Null => { qb.push_bind(None::<String>); },
                    _ => { qb.push_bind(val.to_string()); },
                }
            }
        }

        // Order
        if !self.order_by.is_empty() {
            qb.push(" ORDER BY ");
            let orders: Vec<String> = self.order_by.iter()
                .map(|(c, d)| format!("{} {}", c, d))
                .collect();
            qb.push(orders.join(", "));
        }

        // Limit
        if let Some(l) = self.limit {
            qb.push(" LIMIT ");
            qb.push_bind(l);
        }

        // Offset
        if let Some(o) = self.offset {
            qb.push(" OFFSET ");
            qb.push_bind(o);
        }

        qb
    }
}
