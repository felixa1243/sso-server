'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { get, del, put } from '../../utils/api';

interface Role {
    Name: string;
}

interface User {
    ID: string;
    Email: string;
    Role: Role[];
    IsBanned: boolean;
}

export default function AdminPage() {
    const [users, setUsers] = useState<User[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const router = useRouter();

    useEffect(() => {
        const token = localStorage.getItem('token');
        if (!token) {
            router.push('/login');
            return;
        }

        get('/admin/users', token)
            .then(data => {
                setUsers(data);
                setLoading(false);
            })
            .catch(err => {
                console.error(err);
                setError(err.message);
                setLoading(false);
            });
    }, [router]);

    const handleDelete = async (id: string) => {
        if (!confirm('Are you sure you want to delete this user?')) return;
        const token = localStorage.getItem('token') || '';
        try {
            await del(`/admin/users/${id}`, token);
            setUsers(users.filter(u => u.ID !== id));
        } catch (err: any) {
            alert(err.message);
        }
    };

    const handleBan = async (id: string, isBanned: boolean) => {
        const token = localStorage.getItem('token') || '';
        const endpoint = isBanned ? `/admin/users/${id}/unban` : `/admin/users/${id}/ban`;
        try {
            await put(endpoint, token);
            setUsers(users.map(u => u.ID === id ? { ...u, IsBanned: !isBanned } : u));
        } catch (err: any) {
            alert(err.message);
        }
    };

    if (loading) return <div className="p-4 text-center">Loading...</div>;
    if (error) return <div className="p-4 text-red-500 text-center">Error: {error}</div>;

    return (
        <div className="p-8 bg-gray-50 min-h-screen text-black">
            <div className="max-w-4xl mx-auto">
                <div className="flex justify-between items-center mb-6">
                    <h1 className="text-2xl font-bold">User Management</h1>
                    <div className="flex gap-2">
                        <button
                            onClick={() => router.push('/admin/roles')}
                            className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition"
                        >
                            Manage Roles
                        </button>
                        <button
                            onClick={() => router.push('/admin/clients')}
                            className="px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600 transition"
                        >
                            OAuth Clients
                        </button>
                        <button onClick={() => {
                            localStorage.removeItem('token');
                            localStorage.removeItem('user');
                            router.push('/login');
                        }} className="px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600 transition">Logout</button>
                    </div>
                </div>

                <div className="bg-white shadow rounded-lg overflow-hidden">
                    <table className="min-w-full leading-normal">
                        <thead>
                            <tr>
                                <th className="px-5 py-3 border-b-2 border-gray-200 bg-gray-100 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">
                                    ID
                                </th>
                                <th className="px-5 py-3 border-b-2 border-gray-200 bg-gray-100 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">
                                    Email
                                </th>
                                <th className="px-5 py-3 border-b-2 border-gray-200 bg-gray-100 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">
                                    Roles
                                </th>
                                <th className="px-5 py-3 border-b-2 border-gray-200 bg-gray-100 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">
                                    Status
                                </th>
                                <th className="px-5 py-3 border-b-2 border-gray-200 bg-gray-100 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">
                                    Actions
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            {users.map(user => (
                                <tr key={user.ID}>
                                    <td className="px-5 py-5 border-b border-gray-200 bg-white text-sm">
                                        <p className="text-gray-900 whitespace-no-wrap font-mono text-xs">{user.ID}</p>
                                    </td>
                                    <td className="px-5 py-5 border-b border-gray-200 bg-white text-sm">
                                        <p className="text-gray-900 whitespace-no-wrap">{user.Email}</p>
                                    </td>
                                    <td className="px-5 py-5 border-b border-gray-200 bg-white text-sm">
                                        <span className="relative inline-block px-3 py-1 font-semibold text-green-900 leading-tight">
                                            <span aria-hidden className="absolute inset-0 bg-green-200 opacity-50 rounded-full"></span>
                                            <span className="relative">
                                                {user.Role && user.Role.map(r => r.Name).join(', ')}
                                            </span>
                                        </span>
                                    </td>
                                    <td className="px-5 py-5 border-b border-gray-200 bg-white text-sm">
                                        {user.IsBanned ? (
                                            <span className="text-red-500 font-bold">Banned</span>
                                        ) : (
                                            <span className="text-green-500">Active</span>
                                        )}
                                    </td>
                                    <td className="px-5 py-5 border-b border-gray-200 bg-white text-sm">
                                        <button
                                            onClick={() => router.push(`/admin/punishments/detail?id=${user.ID}`)}
                                            className="text-indigo-600 hover:text-indigo-900 mr-2"
                                        >
                                            Punish
                                        </button>
                                        <button
                                            onClick={() => handleBan(user.ID, user.IsBanned)}
                                            className={`mr-2 ${user.IsBanned ? 'text-green-600 hover:text-green-900' : 'text-orange-600 hover:text-orange-900'}`}
                                        >
                                            {user.IsBanned ? 'Unban' : 'Ban'}
                                        </button>
                                        <button
                                            onClick={() => handleDelete(user.ID)}
                                            className="text-red-600 hover:text-red-900"
                                        >
                                            Delete
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
}
