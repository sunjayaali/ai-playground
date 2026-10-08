// Tokens never cross into JavaScript: the backend
// sets them as httpOnly cookies, so the frontend
// has no token types at all.
export interface UserInfo {
  id: string;
}

export interface AuthError {
  error: string;
}
