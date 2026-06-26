import { useState, useEffect } from 'react';
import './App.css';
import {
    ScanLocalMods,
    GetScannedMods,
    EnableMod,
    DisableMod,
    ImportModFile,
    GetRawDownloads
} from '../wailsjs/go/main/App';

function App() {
    const [mods, setMods] = useState([]);
    const [rawFiles, setRawFiles] = useState([]); // State for unprocessed files
    const [loading, setLoading] = useState(false);
    const [status, setStatus] = useState("");
    const [importPath, setImportPath] = useState("");

    const gamePaksDir = "/home/jonasyke/workspace/github.com/jonasyke/test_mods";

    const fetchData = async () => {
        try {
            const inventoryData = await GetScannedMods();
            setMods(inventoryData || []);

            const rawData = await GetRawDownloads();
            setRawFiles(rawData || []);
        } catch (err) {
            setStatus("Error loading dashboard data: " + err);
        }
    };

    useEffect(() => {
        fetchData();
    }, []);

    const handleScan = async () => {
        setLoading(true);
        setStatus("Scanning...");
        try {
            const count = await ScanLocalMods(gamePaksDir);
            setStatus(`Scan complete! Found and tracked ${count} mod files.`);
            await fetchData();
        } catch (err) {
            setStatus("Error scanning: " + err);
        } finally {
            setLoading(false);
        }
    };

    // Handles manual full-path typing
    const handleManualImport = async (e) => {
        e.preventDefault();
        if (!importPath.trim()) return;

        setStatus("Importing mod file...");
        try {
            const filename = await ImportModFile(importPath.trim());
            setStatus(`Successfully imported: ${filename}`);
            setImportPath("");
            await fetchData();
        } catch (err) {
            setStatus(`Import failed: ${err}`);
        }
    };

    // Quick import logic for files ALREADY inside rawdownloads folder
    const handleQuickImport = async (filename) => {
        setStatus(`Importing ${filename} from cache...`);
        // We construct the path using Flux's expected location pattern
        const fullPath = `/home/jonasyke/workspace/github.com/jonasyke/flux/flux/downloads/cache/rawdownloads/${filename}`;
        try {
            await ImportModFile(fullPath);
            setStatus(`Successfully imported and tracked: ${filename}`);
            await fetchData(); // Refreshes both arrays
        } catch (err) {
            setStatus(`Import failed: ${err}`);
        }
    };

    const handleToggleMod = async (mod, isCurrentlyEnabled) => {
        setStatus(`Processing ${mod.filename}...`);
        try {
            if (isCurrentlyEnabled) {
                await DisableMod(mod.id, mod.filename, gamePaksDir);
                setStatus(`Quarantined: ${mod.filename}`);
            } else {
                await EnableMod(mod.id, mod.filename, gamePaksDir);
                setStatus(`Activated: ${mod.filename}`);
            }
            await fetchData();
        } catch (err) {
            setStatus(`Action failed: ${err}`);
        }
    };

    return (
        <div id="App" style={{ padding: '30px', color: '#fff', background: '#121212', minHeight: '100vh', fontFamily: 'Arial, sans-serif' }}>
            <header style={{ marginBottom: '30px' }}>
                <h1 style={{ margin: 0, color: '#4CAF50' }}>Flux Mod Manager</h1>
                <p style={{ color: '#aaa' }}>Ready or Not Mod Dashboard</p>

                <button
                    onClick={handleScan}
                    disabled={loading}
                    style={{
                        padding: '10px 20px',
                        fontSize: '14px',
                        fontWeight: 'bold',
                        background: '#4CAF50',
                        color: 'white',
                        border: 'none',
                        borderRadius: '4px',
                        cursor: 'pointer'
                    }}
                >
                    {loading ? "Scanning Directories..." : "Scan Mod Directory"}
                </button>

                {status && <p style={{ fontStyle: 'italic', color: '#ffeb3b', marginTop: '15px' }}>{status}</p>}
            </header>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 2fr', gap: '30px' }}>
                {/* LEFT COLUMN: RAW INBOX */}
                <section style={{ background: '#1a1a1a', padding: '20px', borderRadius: '6px', border: '1px solid #2d2d2d' }}>
                    <h3 style={{ margin: '0 0 15px 0', color: '#2196F3', borderBottom: '1px solid #333', paddingBottom: '8px' }}>
                        Raw Downloads Inbox ({rawFiles.length})
                    </h3>
                    {rawFiles.length === 0 ? (
                        <p style={{ color: '#666', fontSize: '13px' }}>No raw .pak files waiting in rawdownloads folder.</p>
                    ) : (
                        <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                            {rawFiles.map((filename) => (
                                <li key={filename} style={{
                                    background: '#252525',
                                    padding: '10px',
                                    borderRadius: '4px',
                                    marginBottom: '10px',
                                    display: 'flex',
                                    flexDirection: 'column',
                                    gap: '8px'
                                }}>
                                    <span style={{ fontSize: '13px', wordBreak: 'break-all', color: '#e0e0e0' }}>{filename}</span>
                                    <button
                                        onClick={() => handleQuickImport(filename)}
                                        style={{
                                            padding: '6px',
                                            background: '#2196F3',
                                            color: '#fff',
                                            border: 'none',
                                            borderRadius: '4px',
                                            cursor: 'pointer',
                                            fontSize: '11px',
                                            fontWeight: 'bold'
                                        }}
                                    >
                                        Process & Import
                                    </button>
                                </li>
                            ))}
                        </ul>
                    )}

                    {/* Manual Fallback Form */}
                    <form onSubmit={handleManualImport} style={{ marginTop: '25px', paddingTop: '15px', borderTop: '1px solid #333' }}>
                        <label style={{ display: 'block', marginBottom: '8px', fontSize: '12px', color: '#888' }}>Import from custom location:</label>
                        <input
                            type="text"
                            placeholder="/absolute/path/to/mod.pak"
                            value={importPath}
                            onChange={(e) => setImportPath(e.target.value)}
                            style={{ width: '100%', padding: '6px', borderRadius: '4px', border: '1px solid #444', background: '#2d2d2d', color: '#fff', boxSizing: 'border-box', marginBottom: '8px', fontSize: '12px' }}
                        />
                        <button type="submit" style={{ width: '100%', padding: '6px', background: '#333', border: '1px solid #555', borderRadius: '4px', color: '#fff', fontSize: '12px', cursor: 'pointer' }}>
                            Import Custom Path
                        </button>
                    </form>
                </section>

                {/* RIGHT COLUMN: MANAGED INVENTORY */}
                <section style={{ background: '#1a1a1a', padding: '20px', borderRadius: '6px', border: '1px solid #2d2d2d' }}>
                    <h3 style={{ margin: '0 0 15px 0', color: '#4CAF50', borderBottom: '1px solid #333', paddingBottom: '8px' }}>
                        Managed Mod Inventory ({mods.length})
                    </h3>
                    {mods.length === 0 ? (
                        <p style={{ color: '#666' }}>No processed mods found. Import a file from the inbox to stage it.</p>
                    ) : (
                        <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left' }}>
                            <thead>
                                <tr style={{ borderBottom: '2px solid #333', color: '#888', fontSize: '13px' }}>
                                    <th style={{ padding: '10px' }}>Filename</th>
                                    <th style={{ padding: '10px' }}>Status</th>
                                    <th style={{ padding: '10px' }}>Action</th>
                                </tr>
                            </thead>
                            <tbody>
                                {mods.map((mod) => {
                                    const isEnabled = !mod.file_path.includes("storage/mods");
                                    return (
                                        <tr key={mod.id} style={{ borderBottom: '1px solid #222', fontSize: '13px' }}>
                                            <td style={{ padding: '10px', color: '#e0e0e0' }}>{mod.filename}</td>
                                            <td style={{ padding: '10px' }}>
                                                <span style={{
                                                    background: isEnabled ? '#2e7d32' : '#c62828',
                                                    padding: '3px 8px',
                                                    borderRadius: '10px',
                                                    fontSize: '10px',
                                                    fontWeight: 'bold'
                                                }}>
                                                    {isEnabled ? "ACTIVE" : "QUARANTINED"}
                                                </span>
                                            </td>
                                            <td style={{ padding: '10px' }}>
                                                <button
                                                    onClick={() => handleToggleMod(mod, isEnabled)}
                                                    style={{ padding: '4px 8px', background: '#333', color: '#fff', border: '1px solid #555', borderRadius: '4px', cursor: 'pointer', fontSize: '11px' }}
                                                >
                                                    {isEnabled ? "Quarantine" : "Activate"}
                                                </button>
                                            </td>
                                        </tr>
                                    );
                                })}
                            </tbody>
                        </table>
                    )}
                </section>
            </div>
        </div>
    );
}

export default App;
