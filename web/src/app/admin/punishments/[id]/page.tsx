'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { get, put, del, post } from '../../../../utils/api';

interface Punishment {
    ID: string;
    Type: string;
    Reason: string;
    StartTime: string;
    EndTime: string;
    AdminID: string;
}

export default function PunishmentsPage({ params }: { params: { id: string } }) {
    const [punishments, setPunishments] = useState<Punishment[]>([]);
    const [loading, setLoading] = useState(true);
    const [newPunishment, setNewPunishment] = useState({
        type: 'BAN',
        reason: '',
        duration_sec: 3600 // 1 hour default
    });
    const router = useRouter();

    useEffect(() => {
        const token = localStorage.getItem('token');
        if (!token) {
            router.push('/login');
            return;
        }

        get(`/admin/users/${params.id}/punishments`, token)
            .then(data => {
                setPunishments(data);
                setLoading(false);
            })
            .catch(err => {
                console.error(err);
                setLoading(false);
            });
    }, [params.id, router]);

    const handlePunish = async (e: React.FormEvent) => {
        e.preventDefault();
        const token = localStorage.getItem('token') || '';
        try {
            // Need to pass body correctly
            const body = {
                user_id: params.id,
                type: newPunishment.type,
                reason: newPunishment.reason,
                duration_sec: Number(newPunishment.duration_sec)
            };

            const result = await post('/admin/punishments', token, body);
            setPunishments([...punishments, result]);
        } catch (err: any) {
            alert(err.message);
        }
    };

    if (loading) return <div>Loading...</div>;

    return (
        <div className="p-8 bg-gray-50 min-h-screen text-black">
            <h1 className="text-2xl font-bold mb-4">User Punishments</h1>

            <div className="bg-white p-6 rounded shadow mb-6">
                <h2 className="text-lg font-semibold mb-4">Add Punishment</h2>
                <form onSubmit={handlePunish} className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700">Type</label>
                        <select
                            value={newPunishment.type}
                            onChange={e => setNewPunishment({...newPunishment, type: e.target.value})}
                            className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2 text-black"
                        >
                            <option value="BAN">Ban</option>
                            <option value="MUTE">Mute</option>
                        </select>
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700">Reason</label>
                        <input
                            type="text"
                            value={newPunishment.reason}
                            onChange={e => setNewPunishment({...newPunishment, reason: e.target.value})}
                            className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2 text-black"
                            required
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700">Duration (seconds)</label>
                        <input
                            type="number"
                            value={newPunishment.duration_sec}
                            onChange={e => setNewPunishment({...newPunishment, duration_sec: parseInt(e.target.value)})}
                            className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2 text-black"
                        />
                    </div>
                    <button type="submit" className="bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700 transition">
                        Apply Punishment
                    </button>
                </form>
            </div>

            <div className="bg-white shadow overflow-hidden sm:rounded-lg">
                <ul className="divide-y divide-gray-200">
                    {punishments.map(p => (
                        <li key={p.ID} className="px-4 py-4 sm:px-6 hover:bg-gray-50">
                            <div className="flex items-center justify-between">
                                <p className="text-sm font-medium text-indigo-600 truncate">{p.Type}</p>
                                <div className="ml-2 flex-shrink-0 flex">
                                    <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${new Date(p.EndTime) > new Date() ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}`}>
                                        {new Date(p.EndTime) > new Date() ? 'Active' : 'Expired'}
                                    </span>
                                </div>
                            </div>
                            <div className="mt-2 sm:flex sm:justify-between">
                                <div className="sm:flex">
                                    <p className="flex items-center text-sm text-gray-500">
                                        Reason: {p.Reason}
                                    </p>
                                </div>
                                <div className="mt-2 flex items-center text-sm text-gray-500 sm:mt-0">
                                    <p>
                                        Ends: {new Date(p.EndTime).toLocaleString()}
                                    </p>
                                </div>
                            </div>
                        </li>
                    ))}
                </ul>
            </div>
        </div>
    );
}
