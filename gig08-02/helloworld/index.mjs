import express from "express";
import pg from "pg";
import { Connector } from "@google-cloud/cloud-sql-connector";

const { Pool } = pg;

const connector = new Connector();
const clientOpts = await connector.getOptions({
  instanceConnectionName: process.env.INSTANCE_CONNECTION_NAME,
  authType: "IAM",
});

const pool = new Pool({
  ...clientOpts,
  user: process.env.DB_USER,
  database: process.env.DB_NAME,
});

const app = express();

app.get("/", async (req, res) => {
  await pool.query("INSERT INTO visits(created_at) VALUES(NOW())");
  const { rows } = await pool.query(
    "SELECT created_at FROM visits ORDER BY created_at DESC LIMIT 5"
  );
  console.table(rows); // prints the last 5 visits
  res.send(rows);
});

const port = parseInt(process.env.PORT) || 8080;
app.listen(port, async () => {
  console.log("process.env: ", process.env);
  await pool.query(`CREATE TABLE IF NOT EXISTS visits (
    id SERIAL NOT NULL,
    created_at timestamp NOT NULL,
    PRIMARY KEY (id)
  );`);
  console.log(`helloworld: listening on port ${port}`);
});
