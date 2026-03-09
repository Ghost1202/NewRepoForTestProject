import { useQuery } from "@tanstack/react-query";
import { getWallet } from "./api";
import { WalletResponse } from "@/entities/wallet";

export const useWallet = () => {
  return useQuery<WalletResponse>({
    queryKey: ["wallet"],
    queryFn: getWallet,
  });
};

