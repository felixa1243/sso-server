'use client';

import { useState, useEffect } from 'react';
import { get, post } from '../../../utils/api';

interface Domain {
    ID: string;
    Name: string;
    URL: string;
    CreatedAt: string;
}

export default function DomainsPage() {
    const [domains, setDomains] = useState<Domain[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    // Form state
    const [name, setName] = useState('');
    const [url, setUrl] = useState('');
    const [submitting, setSubmitting] = useState(false);

    useEffect(() => {
        const token = localStorage.getItem('token');
        if (!token) return;

        get('/domains', token)
            .then((data: any) => {
                setDomains(data || []);
                setLoading(false);
            })
            .catch((err: any) => {
                console.error(err);
                setError(err.message || 'Failed to load domains');
                setLoading(false);
            });
    }, []);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setSubmitting(true);
        setError('');

        const token = localStorage.getItem('token');
        if (!token) {
            setError('Not authenticated');
            setSubmitting(false);
            return;
        }

        try {
            const newDomain = await post('/domains', { name, url }, token);
            setDomains([...domains, newDomain]);
            setName('');
            setUrl('');
        } catch (err: any) {
            setError(err.message || 'Failed to create domain');
        } finally {
            setSubmitting(false);
        }
    };

    if (loading) return <div className="p-4 text-center text-gray-600">Loading domains...</div>;

    return (
        <div className="space-y-6 max-w-4xl mx-auto">
            <h1 className="text-3xl font-bold text-gray-800">Domain Management</h1>

            {/* Registration Form */}
            <div className="bg-white p-6 rounded-lg shadow-md border border-gray-200">
                <h2 className="text-xl font-semibold mb-4 text-gray-800">Register New Domain</h2>
                {error && <div className="p-3 mb-4 text-red-700 bg-red-100 rounded-md border border-red-200">{error}</div>}

                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">Domain Name</label>
                        <input
                            type="text"
                            id="name"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            placeholder="e.g., mysite"
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
                            required
                        />
                        <p className="text-xs text-gray-500 mt-1">Unique identifier for your domain.</p>
                    </div>
                    <div>
                        <label htmlFor="url" className="block text-sm font-medium text-gray-700 mb-1">Domain URL</label>
                        <input
                            type="url"
                            id="url"
                            value={url}
                            onChange={(e) => setUrl(e.target.value)}
                            placeholder="https://mysite.com"
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
                            required
                        />
                        <p className="text-xs text-gray-500 mt-1">Full URL of your application.</p>
                    </div>
                    <button
                        type="submit"
                        disabled={submitting}
                        className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-sm font-medium"
                    >
                        {submitting ? 'Registering...' : 'Register Domain'}
                    </button>
                </form>
            </div>

            {/* List Domains */}
            <div className="bg-white p-6 rounded-lg shadow-md border border-gray-200">
                <h2 className="text-xl font-semibold mb-4 text-gray-800">Your Domains</h2>
                {domains.length === 0 ? (
                    <div className="text-center py-8 bg-gray-50 rounded-lg border border-dashed border-gray-300">
                        <p className="text-gray-500 italic">No domains registered yet.</p>
                        <p className="text-sm text-gray-400 mt-1">Use the form above to register your first domain.</p>
                    </div>
                ) : (
                    <div className="overflow-x-auto rounded-lg border border-gray-200">
                        <table className="min-w-full divide-y divide-gray-200">
                            <thead className="bg-gray-50">
                                <tr>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">URL</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created At</th>
                                </tr>
                            </thead>
                            <tbody className="bg-white divide-y divide-gray-200">
                                {domains.map((domain) => (
                                    <tr key={domain.ID} className="hover:bg-gray-50 transition-colors">
                                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{domain.Name}</td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                            <a href={domain.URL} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:text-blue-800 hover:underline inline-flex items-center gap-1">
                                                {domain.URL}
                                                <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"></path></svg>
                                            </a>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                            {new Date(domain.CreatedAt).toLocaleDateString()}
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
