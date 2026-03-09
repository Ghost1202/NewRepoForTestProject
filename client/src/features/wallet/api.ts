import { walletApi } from "@/shared/api";
import { WalletResponse } from "@/entities/wallet";

export const getWallet = async () => {
  const response = await walletApi.get<WalletResponse>("/wallet", { authMode: "required" });
  return response.data;
};

