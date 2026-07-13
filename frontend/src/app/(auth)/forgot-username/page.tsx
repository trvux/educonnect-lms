import { ForgotUsernameForm } from "./forgot-username-form";

export default function ForgotUsernamePage() {
  return (
    <div className="flex min-h-svh w-full items-center justify-center p-4 sm:p-6 md:p-10">
      <div className="w-full max-w-sm">
        <ForgotUsernameForm />
      </div>
    </div>
  );
}
