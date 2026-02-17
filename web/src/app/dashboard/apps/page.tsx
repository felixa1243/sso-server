'use client';

import { useState, useEffect } from 'react';
import { get, post } from '../../../utils/api';

interface ClientApp {
    id: string;
    name: string;
    client_id: string;
    redirect_uris: string;
    scopes: string;
    // client_secret is usually not returned in list
}

interface NewClientApp extends ClientApp {
    client_secret: string;
}

export default function AppsPage() {
    const [apps, setApps] = useState<ClientApp[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [createdApp, setCreatedApp] = useState<NewClientApp | null>(null);

    // Form state
    const [name, setName] = useState('');
    const [redirectUris, setRedirectUris] = useState('');
    const [scopes, setScopes] = useState('openid profile email');
    const [submitting, setSubmitting] = useState(false);

    useEffect(() => {
        fetchApps();
    }, []);

    const fetchApps = () => {
        const token = localStorage.getItem('token');
        if (!token) return;

        get('/clients', token)
            .then((data: any) => {
                setApps(data || []);
                setLoading(false);
            })
            .catch((err: any) => {
                console.error(err);
                setError(err.message || 'Failed to load apps');
                setLoading(false);
            });
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setSubmitting(true);
        setError('');
        setCreatedApp(null);

        const token = localStorage.getItem('token');
        if (!token) {
            setError('Not authenticated');
            setSubmitting(false);
            return;
        }

        try {
            const newApp = await post('/clients', {
                name,
                redirect_uris: redirectUris,
                scopes
            }, token);

            setCreatedApp(newApp);
            // Add to list but without secret (or just refresh list)
            // The response has the secret, so we should be careful not to expose it in the list if we just push it.
            // But for UI state, it's fine.
            setApps([...apps, newApp]);

            // Clear form
            setName('');
            setRedirectUris('');
            setScopes('openid profile email');
        } catch (err: any) {
            setError(err.message || 'Failed to create app');
        } finally {
            setSubmitting(false);
        }
    };

    if (loading) return <div className="p-4 text-center text-gray-600">Loading apps...</div>;

    return (
        <div className="space-y-6 max-w-4xl mx-auto">
            <h1 className="text-3xl font-bold text-gray-800">OAuth Apps Management</h1>

            {/* Registration Form */}
            <div className="bg-white p-6 rounded-lg shadow-md border border-gray-200">
                <h2 className="text-xl font-semibold mb-4 text-gray-800">Register New App</h2>
                {error && <div className="p-3 mb-4 text-red-700 bg-red-100 rounded-md border border-red-200">{error}</div>}

                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">App Name</label>
                        <input
                            type="text"
                            id="name"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            placeholder="My Awesome App"
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
                            required
                        />
                    </div>
                    <div>
                        <label htmlFor="redirectUris" className="block text-sm font-medium text-gray-700 mb-1">Redirect URIs</label>
                        <input
                            type="text"
                            id="redirectUris"
                            value={redirectUris}
                            onChange={(e) => setRedirectUris(e.target.value)}
                            placeholder="https://myapp.com/callback, http://localhost:3000/callback"
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
                            required
                        />
                        <p className="text-xs text-gray-500 mt-1">Comma-separated list of allowed callback URLs.</p>
                    </div>
                    <div>
                        <label htmlFor="scopes" className="block text-sm font-medium text-gray-700 mb-1">Scopes</label>
                        <input
                            type="text"
                            id="scopes"
                            value={scopes}
                            onChange={(e) => setScopes(e.target.value)}
                            placeholder="openid profile email"
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
                            required
                        />
                        <p className="text-xs text-gray-500 mt-1">Space-separated list of scopes.</p>
                    </div>
                    <button
                        type="submit"
                        disabled={submitting}
                        className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-sm font-medium"
                    >
                        {submitting ? 'Registering...' : 'Register App'}
                    </button>
                </form>
            </div>

            {/* Success Message with Credentials */}
            {createdApp && (
                <div className="bg-green-50 p-6 rounded-lg border border-green-200 shadow-sm">
                    <h3 className="text-lg font-semibold text-green-800 mb-2">App Created Successfully!</h3>
                    <p className="text-sm text-green-700 mb-4">Please copy your Client Secret now. It will not be shown again.</p>

                    <div className="space-y-3">
                        <div>
                            <span className="block text-xs font-medium text-gray-500 uppercase tracking-wider">Client ID</span>
                            <div className="mt-1 flex items-center bg-white border border-gray-300 rounded-md px-3 py-2">
                                <code className="flex-1 text-sm text-gray-800 break-all">{createdApp.client_id}</code>
                            </div>
                        </div>
                        <div>
                            <span className="block text-xs font-medium text-gray-500 uppercase tracking-wider">Client Secret</span>
                            <div className="mt-1 flex items-center bg-white border border-red-200 rounded-md px-3 py-2">
                                <code className="flex-1 text-sm text-red-600 font-bold break-all">{createdApp.client_secret}</code>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {/* List Apps */}
            <div className="bg-white p-6 rounded-lg shadow-md border border-gray-200">
                <h2 className="text-xl font-semibold mb-4 text-gray-800">Your Apps</h2>
                {apps.length === 0 ? (
                    <div className="text-center py-8 bg-gray-50 rounded-lg border border-dashed border-gray-300">
                        <p className="text-gray-500 italic">No apps registered yet.</p>
                    </div>
                ) : (
                    <div className="overflow-x-auto rounded-lg border border-gray-200">
                        <table className="min-w-full divide-y divide-gray-200">
                            <thead className="bg-gray-50">
                                <tr>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Client ID</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Redirect URIs</th>
                                </tr>
                            </thead>
                            <tbody className="bg-white divide-y divide-gray-200">
                                {apps.map((app) => (
                                    <tr key={app.id} className="hover:bg-gray-50 transition-colors">
                                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{app.name}</td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 font-mono text-xs">{app.client_id}</td>
                                        <td className="px-6 py-4 text-sm text-gray-500 max-w-xs truncate" title={app.redirect_uris}>
                                            {app.redirect_uris}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>
        </div>
    );
}
