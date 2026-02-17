'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { login } from '../../utils/api';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [domain, setDomain] = useState('');
  const [error, setError] = useState('');
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const data = await login(email, password, domain || undefined);
      localStorage.setItem('token', data.access_token);
      localStorage.setItem('user', JSON.stringify(data.user));
      router.push('/admin');
    } catch (err: any) {
      setError(err.message || 'Login failed');
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-100">
      <form onSubmit={handleSubmit} className="p-6 bg-white rounded shadow-md w-80 text-black">
        <h1 className="mb-4 text-xl font-bold">Login</h1>
        {error && <p className="mb-2 text-red-500 text-sm">{error}</p>}
        <div className="mb-4">
            <label className="block mb-1 text-sm font-semibold">Email</label>
            <input
              type="email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              className="w-full p-2 border rounded text-black"
              required
            />
        </div>
        <div className="mb-4">
            <label className="block mb-1 text-sm font-semibold">Password</label>
            <input
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              className="w-full p-2 border rounded text-black"
              required
            />
        </div>
        <div className="mb-4">
            <label className="block mb-1 text-sm font-semibold">Domain (Optional)</label>
            <input
              type="text"
              value={domain}
              onChange={e => setDomain(e.target.value)}
              className="w-full p-2 border rounded text-black"
              placeholder="e.g. blog"
            />
        </div>
        <button type="submit" className="w-full p-2 text-white bg-blue-500 rounded hover:bg-blue-600 transition">
          Sign In
        </button>
      </form>
    </div>
  );
}
