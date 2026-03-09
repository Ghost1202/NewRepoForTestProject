import { BrowserRouter } from "react-router-dom";
import { AuthProvider, QueryProvider } from "./providers";
import { AppRouter } from "./router";

export const App = () => {
  return (
    <BrowserRouter>
      <QueryProvider>
        <AuthProvider>
          <AppRouter />
        </AuthProvider>
      </QueryProvider>
    </BrowserRouter>
  );
};
