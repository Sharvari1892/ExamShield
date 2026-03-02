import { createBrowserRouter, Navigate } from "react-router-dom";
import { LoginPage } from "@/features/auth/LoginPage";
import { ExamPage } from "@/features/exam/ExamPage";
import { PrivateRoute } from "./PrivateRoute";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <Navigate to="/login" replace />,
  },
  {
    path: "/login",
    element: <LoginPage />,
  },
  {
    path: "/exam",
    element: (
      <PrivateRoute>
        <ExamPage />
      </PrivateRoute>
    ),
  },
]);