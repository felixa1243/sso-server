'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { get, post, del } from '../../../utils/api';

interface ClientApp {
    ID: string;
    Name: string;
    RedirectURIs: string;
    AllowedScopes: string;
    Secret?: string; // Only shown on creation
}

export default function ClientsPage() {
    const [clients, setClients] = useState<ClientApp[]>([]);
    const [loading, setLoading] = useState(true);
    const [newClient, setNewClient] = useState({
        name: '',
        redirect_uris: '',
        scopes: ''
    });
    const [createdSecret, setCreatedSecret] = useState('');
    const router = useRouter();

    useEffect(() => {
        const token = localStorage.getItem('token');
        if (!token) {
            router.push('/login');
            return;
        }

        get('/admin/clients', token)
            .then(data => {
                setClients(data);
                setLoading(false);
            })
            .catch(err => {
                console.error(err);
                setLoading(false);
            });
    }, [router]);

    const handleCreate = async () => {
        const token = localStorage.getItem('token') || '';
        try {
            const result = await post('/admin/clients', token, newClient);
            // Result should contain client object and 'secret' (plain text)
            if (result.secret) {
                setCreatedSecret(result.secret);
                alert(`Client created! SECRET: ${result.secret} (Save this now!)`);
            }
            // Refresh list
            const updated = await get('/admin/clients', token);
            setClients(updated);
            setNewClient({ name: '', redirect_uris: '', scopes: '' });
        } catch (err: any) {
            alert(err.message);
        }
    };

    const handleDelete = async (id: string) => {
        if (!confirm('Delete this client app?')) return;
        const token = localStorage.getItem('token') || '';
        try {
            await del(`/admin/clients/${id}`, token);
            setClients(clients.filter(c => c.ID !== id));
        } catch (err: any) {
            alert(err.message);
        }
    };

    if (loading) return <div>Loading...</div>;

    return (
        <div className="p-8 bg-gray-50 min-h-screen text-black">
            <h1 className="text-2xl font-bold mb-6">OAuth Consumers (Clients)</h1>

            {createdSecret && (
                <div className="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded relative mb-4">
                    <strong className="font-bold">New Client Secret:</strong>
                    <span className="block sm:inline ml-2 font-mono">{createdSecret}</span>
                    <p className="text-sm mt-1">Please copy this secret. It will not be shown again.</p>
                    <button onClick={() => setCreatedSecret('')} className="absolute top-0 bottom-0 right-0 px-4 py-3">
                        <span className="text-green-500">×</span>
                    </button>
                </div>
            )}

            <div className="bg-white p-6 rounded shadow mb-8">
                <h2 className="text-lg font-semibold mb-4">Register New Client</h2>
                <div className="grid grid-cols-1 gap-4">
                    <input
                        className="border p-2 rounded text-black"
                        placeholder="App Name"
                        value={newClient.name}
                        onChange={e => setNewClient({...newClient, name: e.target.value})}
                    />
                    <input
                        className="border p-2 rounded text-black"
                        placeholder="Redirect URIs (comma separated)"
                        value={newClient.redirect_uris}
                        onChange={e => setNewClient({...newClient, redirect_uris: e.target.value})}
                    />
                    <input
                        className="border p-2 rounded text-black"
                        placeholder="Scopes (space separated)"
                        value={newClient.scopes}
                        onChange={e => setNewClient({...newClient, scopes: e.target.value})}
                    />
                    <button onClick={handleCreate} className="bg-blue-600 text-white p-2 rounded hover:bg-blue-700">
                        Register Client
                    </button>
                </div>
            </div>

            <div className="grid gap-6">
                {clients.map(client => (
                    <div key={client.ID} className="bg-white p-4 rounded shadow border border-gray-200">
                        <div className="flex justify-between items-start">
                            <div>
                                <h3 className="font-bold text-lg">{client.Name}</h3>
                                <p className="text-sm text-gray-600 font-mono">ID: {client.ID}</p>
                            </div>
                            <button onClick={() => handleDelete(client.ID)} className="text-red-600 text-sm hover:underline">
                                Delete
                            </button>
                        </div>
                        <div className="mt-4 text-sm">
                            <p><strong>Redirect URIs:</strong> {client.RedirectURIs}</p>
                            <p><strong>Scopes:</strong> {client.AllowedScopes}</p>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}
