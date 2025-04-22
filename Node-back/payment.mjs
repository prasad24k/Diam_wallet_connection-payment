import express from 'express';
import cors from 'cors';
import {
  Asset,
  Aurora,
  BASE_FEE,
  Operation,
  TransactionBuilder,
} from "diamnet-sdk";

const app = express();
const port = 3000;

const NETWORK_PASSPHRASE = "Diamante Testnet 2024";

app.use(express.json());
app.use(cors()); // Enable CORS for all routes

app.post('/create-transaction', async (req, res) => {
  const { sourcePublicKey, destination, amount } = req.body;

  try {
    const server = new Aurora.Server("https://diamtestnet.diamcircle.io/");

    // Create the transaction
    const account = await server.loadAccount(sourcePublicKey);

    if (!account) {
      return res.status(404).json({ error: "Source account not found" });
    }

    const tx = new TransactionBuilder(account, {
      fee: BASE_FEE,
      networkPassphrase: NETWORK_PASSPHRASE,
    })
      .addOperation(
        Operation.payment({
          destination: destination,
          asset: Asset.native(),
          amount: amount,
        })
      )
      .setTimeout(30)
      .build();

    const transactionXDR = tx.toEnvelope().toXDR("base64");

    res.json({ transactionXDR });
  } catch (error) {
    console.error(`Error creating transaction: ${error.message}`);
    res.status(500).json({ error: error.message });
  }
});

app.listen(port, () => {
  console.log(`API listening at http://localhost:${port}`);
});
