'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { get, put, del, post } from '../../../utils/api';

interface Permission {
    ID: string;
    Name: string;
    Slug: string;
}

interface Role {
    ID: string;
    Name: string;
    Permissions: Permission[];
}

export default function RolesPage() {
    const [roles, setRoles] = useState<Role[]>([]);
    const [permissions, setPermissions] = useState<Permission[]>([]);
    const [loading, setLoading] = useState(true);
    const [newRoleName, setNewRoleName] = useState('');
    const [selectedRole, setSelectedRole] = useState<string | null>(null);
    const [selectedPerms, setSelectedPerms] = useState<Set<string>>(new Set());
    const router = useRouter();

    useEffect(() => {
        const token = localStorage.getItem('token');
        if (!token) {
            router.push('/login');
            return;
        }

        Promise.all([
            get('/admin/roles', token),
            get('/admin/permissions', token)
        ]).then(([rolesData, permsData]) => {
            setRoles(rolesData);
            setPermissions(permsData);
            setLoading(false);
        }).catch(err => {
            alert('Failed to load data: ' + err.message);
            setLoading(false);
        });
    }, [router]);

    const handleCreateRole = async () => {
        const token = localStorage.getItem('token') || '';
        try {
            const role = await post('/admin/roles', token, { name: newRoleName });
            setRoles([...roles, role]);
            setNewRoleName('');
        } catch (err: any) {
            alert(err.message);
        }
    };

    const handleDeleteRole = async (id: string) => {
        if (!confirm('Delete role?')) return;
        const token = localStorage.getItem('token') || '';
        try {
            await del(`/admin/roles/${id}`, token);
            setRoles(roles.filter(r => r.ID !== id));
        } catch (err: any) {
            alert(err.message);
        }
    };

    const handleAssignPermissions = async (roleID: string) => {
        const token = localStorage.getItem('token') || '';
        try {
            await post(`/admin/roles/${roleID}/permissions`, token, { permission_ids: Array.from(selectedPerms) });
            alert('Permissions updated');
            // Reload roles
            const updatedRoles = await get('/admin/roles', token);
            setRoles(updatedRoles);
        } catch (err: any) {
            alert(err.message);
        }
    };

    if (loading) return <div>Loading...</div>;

    return (
        <div className="p-8 bg-gray-50 min-h-screen text-black">
            <h1 className="text-2xl font-bold mb-6">Role Management</h1>

            <div className="mb-8 p-4 bg-white rounded shadow">
                <h2 className="text-lg font-semibold mb-2">Create Role</h2>
                <div className="flex gap-2">
                    <input
                        className="border p-2 rounded"
                        value={newRoleName}
                        onChange={e => setNewRoleName(e.target.value)}
                        placeholder="Role Name"
                    />
                    <button onClick={handleCreateRole} className="bg-blue-500 text-white p-2 rounded">Create</button>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="bg-white p-4 rounded shadow">
                    <h2 className="text-lg font-semibold mb-4">Roles</h2>
                    <ul>
                        {roles.map(role => (
                            <li key={role.ID} className="flex justify-between items-center border-b py-2">
                                <div>
                                    <span className="font-bold">{role.Name}</span>
                                    <p className="text-xs text-gray-500">
                                        {role.Permissions?.map(p => p.Slug).join(', ')}
                                    </p>
                                </div>
                                <div>
                                    <button
                                        onClick={() => {
                                            setSelectedRole(role.ID);
                                            setSelectedPerms(new Set(role.Permissions?.map(p => p.ID) || []));
                                        }}
                                        className="text-blue-600 mr-2"
                                    >
                                        Edit Perms
                                    </button>
                                    <button onClick={() => handleDeleteRole(role.ID)} className="text-red-600">Delete</button>
                                </div>
                            </li>
                        ))}
                    </ul>
                </div>

                {selectedRole && (
                    <div className="bg-white p-4 rounded shadow">
                        <h2 className="text-lg font-semibold mb-4">Edit Permissions for Role</h2>
                        <div className="max-h-96 overflow-y-auto">
                            {permissions.map(perm => (
                                <div key={perm.ID} className="flex items-center mb-2">
                                    <input
                                        type="checkbox"
                                        checked={selectedPerms.has(perm.ID)}
                                        onChange={e => {
                                            const newSet = new Set(selectedPerms);
                                            if (e.target.checked) newSet.add(perm.ID);
                                            else newSet.delete(perm.ID);
                                            setSelectedPerms(newSet);
                                        }}
                                        className="mr-2"
                                    />
                                    <span>{perm.Name} ({perm.Slug})</span>
                                </div>
                            ))}
                        </div>
                        <button onClick={() => handleAssignPermissions(selectedRole)} className="mt-4 bg-green-500 text-white p-2 rounded w-full">
                            Save Permissions
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
}
